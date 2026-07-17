/*
Package storage defines persistence interfaces for campaign history and a
JSON-backed implementation.

Per the remediation plan, SQLite is deferred; the application layer depends only
on these interfaces, so a SQLite (or other) implementation can replace the JSON
store later without rewriting callers.
*/
package storage

import (
	"time"

	"MailMergeApp/backend/campaign"
)

// CampaignRecord is a persisted snapshot of a completed (or terminal) campaign
// run: enough to review it, retry failed recipients, and export results.
type CampaignRecord struct {
	ID             string                  `json:"id"`
	StartedAt      time.Time               `json:"startedAt" ts_type:"string"`
	FinishedAt     time.Time               `json:"finishedAt" ts_type:"string"`
	Subject        string                  `json:"subject"`
	IsHTML         bool                    `json:"isHTML"`
	RecipientCount int                     `json:"recipientCount"`
	State          campaign.CampaignState  `json:"state"`
	Result         campaign.CampaignResult `json:"result"`
}

// CampaignRepository persists and retrieves campaign run records.
type CampaignRepository interface {
	Save(rec CampaignRecord) error
	List() ([]CampaignRecord, error) // newest first
	Get(id string) (*CampaignRecord, error)
	Delete(id string) error
}
