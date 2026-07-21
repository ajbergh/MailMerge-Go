package storage

import (
	"testing"
	"time"

	"MailMergeApp/backend/campaign"
)

func TestJSONRepositorySaveListGet(t *testing.T) {
	repo, err := NewJSONCampaignRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rec := CampaignRecord{
		ID:             "run-1",
		StartedAt:      time.Now().Add(-time.Minute),
		Subject:        "Hello",
		DraftOnly:      true,
		RecipientCount: 3,
		State:          campaign.CampaignCompleted,
		Result:         campaign.CampaignResult{State: campaign.CampaignCompleted, Submitted: 3},
	}
	if err := repo.Save(rec); err != nil {
		t.Fatal(err)
	}
	rec2 := rec
	rec2.ID = "run-2"
	rec2.StartedAt = time.Now()
	if err := repo.Save(rec2); err != nil {
		t.Fatal(err)
	}

	list, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != "run-2" {
		t.Errorf("List should return newest first, got %+v", list)
	}

	got, err := repo.Get("run-1")
	if err != nil || got.Subject != "Hello" {
		t.Errorf("Get failed: %v %+v", err, got)
	}
	if !got.DraftOnly {
		t.Error("draft-only campaign mode was not preserved")
	}

	if err := repo.Delete("run-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get("run-1"); err == nil {
		t.Error("record should be deleted")
	}
}

func TestJSONRepositoryRejectsBadID(t *testing.T) {
	repo, _ := NewJSONCampaignRepository(t.TempDir())
	if err := repo.Save(CampaignRecord{ID: "../evil"}); err == nil {
		t.Error("should reject traversal id")
	}
}
