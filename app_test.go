package main

import (
	"context"
	"testing"

	"MailMergeApp/backend/campaign"
	"MailMergeApp/backend/models"
	"MailMergeApp/backend/storage"
)

func TestRetryCampaignPersistsLineageAcrossRestart(t *testing.T) {
	repo, err := storage.NewJSONCampaignRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	firstSender := campaign.NewFakeSender()
	firstSender.FailFor = map[string]bool{"failed@example.com": true}
	firstApp := &App{
		ctx:     context.Background(),
		sender:  firstSender,
		runner:  campaign.NewRunner(firstSender),
		history: repo,
	}
	request := models.EmailRequest{
		Contacts:        []models.Contact{{ID: "contact-1", FirstName: "Ada", Email: "failed@example.com"}},
		SubjectTemplate: "Hello {{first_name}}",
		BodyTemplate:    "Body",
	}

	initial := firstApp.SendBulkEmails(request)
	if initial.CampaignID == "" {
		t.Fatal("initial run did not receive a persisted campaign ID")
	}
	if initial.State != campaign.CampaignCompletedWithFailures || initial.Failed != 1 {
		t.Fatalf("unexpected initial result: %+v", initial)
	}

	// Simulate an application restart: create a new App with no in-memory lastResult.
	retrySender := campaign.NewFakeSender()
	restartedApp := &App{
		ctx:     context.Background(),
		sender:  retrySender,
		runner:  campaign.NewRunner(retrySender),
		history: repo,
	}
	retried := restartedApp.RetryCampaign(initial.CampaignID)
	if retried.State != campaign.CampaignCompleted || retried.Submitted != 1 || retried.Failed != 0 {
		t.Fatalf("unexpected retry result: %+v", retried)
	}
	if retried.ParentCampaignID != initial.CampaignID || retried.RunNumber != 2 || retried.CampaignID == "" {
		t.Fatalf("retry lineage is incorrect: %+v", retried)
	}
	if len(retried.RecipientResults) != 1 || len(retried.RecipientResults[0].Attempts) != 2 {
		t.Fatalf("retry did not preserve attempt history: %+v", retried.RecipientResults)
	}

	records, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected original and retry records, got %d", len(records))
	}
	child, err := repo.Get(retried.CampaignID)
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentCampaignID != initial.CampaignID || child.RunNumber != 2 {
		t.Fatalf("persisted child lineage is incorrect: %+v", child)
	}
	if child.SubjectTemplate != request.SubjectTemplate || len(child.Contacts) != 1 {
		t.Fatalf("persisted snapshot is incomplete: %+v", child)
	}
}

func TestCampaignSnapshotIsIndependentFromRequestMutation(t *testing.T) {
	repo, err := storage.NewJSONCampaignRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sender := campaign.NewFakeSender()
	app := &App{ctx: context.Background(), sender: sender, runner: campaign.NewRunner(sender), history: repo}
	request := models.EmailRequest{
		Contacts: []models.Contact{{
			ID:           "contact-1",
			Email:        "a@example.com",
			CustomFields: map[string]string{"company": "Original"},
		}},
		SubjectTemplate: "Hello",
		BodyTemplate:    "{{company}}",
	}
	result := app.SendBulkEmails(request)
	request.Contacts[0].Email = "mutated@example.com"
	request.Contacts[0].CustomFields["company"] = "Mutated"

	record, err := repo.Get(result.CampaignID)
	if err != nil {
		t.Fatal(err)
	}
	if record.Contacts[0].Email != "a@example.com" || record.Contacts[0].CustomFields["company"] != "Original" {
		t.Fatalf("persisted snapshot changed with caller mutation: %+v", record.Contacts[0])
	}
}

func TestHeadlessCampaignDoesNotRequireWailsRuntimeContext(t *testing.T) {
	sender := campaign.NewFakeSender()
	app := &App{sender: sender, runner: campaign.NewRunner(sender)}
	result := app.SendBulkEmails(models.EmailRequest{
		Contacts:        []models.Contact{{ID: "headless", Email: "headless@example.com"}},
		SubjectTemplate: "Headless",
		BodyTemplate:    "Body",
	})
	if result.State != campaign.CampaignCompleted || result.Submitted != 1 {
		t.Fatalf("unexpected headless campaign result: %+v", result)
	}
}
