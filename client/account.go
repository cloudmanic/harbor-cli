// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// StartAccountExport starts a full-account GDPR data export. The call is
// idempotent-ish: if a queued/running export job already exists for the user the
// server returns that job instead of starting a new one. Returns the data-wrapped
// export job ({export_job_id, status}); the response status is 202 Accepted.
func (c *Client) StartAccountExport() ([]byte, error) {
	return c.doPost("/account/export", map[string]any{})
}

// GetAccountExport polls an export job by id. When the job is completed and the
// result blob has not expired, the data-wrapped job includes a short-lived
// presigned download_url and a result_expires_at. A failed job carries error_text.
func (c *Client) GetAccountExport(id string) ([]byte, error) {
	return c.doGet("/account/export/"+id, nil)
}

// RequestAccountDeletion schedules a grace-period account deletion. It requires
// re-auth via current_password AND the exact confirmation phrase (sent verbatim
// as confirm). No data is destroyed now: the response carries the scheduled
// status, purge_after, grace_days, and can_cancel_until window.
func (c *Client) RequestAccountDeletion(currentPassword, confirm string) ([]byte, error) {
	return c.doPost("/account/delete", map[string]any{
		"current_password": currentPassword,
		"confirm":          confirm,
	})
}

// CancelAccountDeletion cancels a pending deletion within the grace window,
// re-authenticating via current_password. Returns the data-wrapped reactivated
// status ({status: "active"}).
func (c *Client) CancelAccountDeletion(currentPassword string) ([]byte, error) {
	return c.doPost("/account/delete/cancel", map[string]any{
		"current_password": currentPassword,
	})
}
