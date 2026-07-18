/*
Package storage defines persistence interfaces for campaign history and a
JSON-backed implementation.

SQLite remains a future implementation option; application code depends on this
repository interface so storage can evolve without coupling campaign execution to
a particular database.
*/
package storage

import (
	"time"

	"MailMergeApp/backend/campaign"
	"MailMergeApp/backend/email"
	"MailMergeApp/backend/models"
)

// CampaignRecord is an immutable snapshot of one campaign run. It stores enough
// original input to review a campaign and reconstruct a safe retry after restart.
// Retries are written as new records linked by ParentCampaignID; the original
// record is never overwritten with a retry result.
type CampaignRecord struct {
	ID               string                  `json:"id"`
	ParentCampaignID string                  `json:"parentCampaignId,omitempty"`
	RunNumber        int                     `json:"runNumber"`
	CreatedAt        time.Time               `json:"createdAt" ts_type:"string"`
	StartedAt        time.Time               `json:"startedAt" ts_type:"string"`
	FinishedAt       time.Time               `json:"finishedAt" ts_type:"string"`
	Duration         time.Duration           `json:"duration"`
	Subject          string                  `json:"subject"`
	SubjectTemplate  string                  `json:"subjectTemplate"`
	BodyTemplate     string                  `json:"bodyTemplate"`
	IsHTML           bool                    `json:"isHTML"`
	DraftOnly        bool                    `json:"draftOnly,omitempty"`
	Headers          []string                `json:"headers,omitempty"`
	Contacts         []models.Contact        `json:"contacts,omitempty"`
	Attachments      []string                `json:"attachments,omitempty"`
	CCTemplate       string                  `json:"ccTemplate,omitempty"`
	BCCTemplate      string                  `json:"bccTemplate,omitempty"`
	DuplicatePolicy  email.Policy            `json:"duplicatePolicy"`
	SendOptions      campaign.SendOptions    `json:"sendOptions"`
	SenderType       string                  `json:"senderType"`
	RecipientCount   int                     `json:"recipientCount"`
	State            campaign.CampaignState  `json:"state"`
	Result           campaign.CampaignResult `json:"result"`
}

// CampaignRepository persists and retrieves campaign run snapshots.
type CampaignRepository interface {
	Create(rec CampaignRecord) error
	Update(rec CampaignRecord) error
	Save(rec CampaignRecord) error // compatibility alias for upsert-style callers
	List() ([]CampaignRecord, error)
	Get(id string) (*CampaignRecord, error)
	Delete(id string) error
}
