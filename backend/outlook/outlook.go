package outlook

import "MailMergeApp/backend/campaign"

// Compile-time assurance that the platform Sender implements EmailSender.
var _ campaign.EmailSender = (*Sender)(nil)
