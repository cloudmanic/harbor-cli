// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// GetProfile fetches the authenticated user's profile (wrapped in data). The
// profile id doubles as the sync scope_id.
func (c *Client) GetProfile() ([]byte, error) {
	return c.doGet("/profile", nil)
}

// UpdateProfile applies a partial profile update (name, locale, timezone, or a
// staged email change with current_password) and returns the updated profile.
func (c *Client) UpdateProfile(body map[string]any) ([]byte, error) {
	return c.doPatch("/profile", body)
}

// ChangePassword changes the account password. The server revokes other
// sessions on success.
func (c *Client) ChangePassword(currentPassword, newPassword string) ([]byte, error) {
	return c.doPost("/profile/change-password", map[string]any{
		"current_password": currentPassword,
		"new_password":     newPassword,
	})
}

// SetAvatar points the profile avatar at an already-uploaded blob key.
func (c *Client) SetAvatar(blobKey string) ([]byte, error) {
	return c.doPost("/profile/avatar", map[string]any{"blob_key": blobKey})
}

// RemoveAvatar clears the profile avatar.
func (c *Client) RemoveAvatar() ([]byte, error) {
	return c.doDelete("/profile/avatar", nil)
}

// ConfirmEmailChange consumes a staged email-change confirmation token (public).
func (c *Client) ConfirmEmailChange(token string) ([]byte, error) {
	return c.postJSONNoRefresh("/profile/confirm-email", map[string]string{"token": token})
}
