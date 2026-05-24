package model

// IAMUserWasteInfo holds information about unused IAM users.
type IAMUserWasteInfo struct {
	UserName         string `json:"user_name"`
	PasswordLastUsed string `json:"password_last_used"` // e.g., "Never" or "120 days ago"
	AccessKeysStatus string `json:"access_keys_status"` // e.g., "No active keys" or "All keys unused > 90 days"
}

// IAMRootUserWasteInfo holds information about the root user if MFA is not enabled.
type IAMRootUserWasteInfo struct {
	HasMFA bool `json:"has_mfa"`
}
