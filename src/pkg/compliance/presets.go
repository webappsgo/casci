package compliance

var presetConfigs = map[Mode]*ComplianceConfig{
	ModeHIPAA: {
		Mode:                  ModeHIPAA,
		Enabled:               true,
		RequireMFA:            true,
		MinPasswordLength:     12,
		PasswordExpiry:        90,
		RequirePasswordChange: true,
		MaxFailedLogins:       5,
		SessionTimeout:        15,
		RequireAudit:          true,
		RequireEncryption:     true,
		DataRetention:         2555, // 7 years
		RequireApprovals:      true,
	},
	ModeSOX: {
		Mode:                  ModeSOX,
		Enabled:               true,
		RequireMFA:            true,
		MinPasswordLength:     10,
		PasswordExpiry:        90,
		RequirePasswordChange: true,
		MaxFailedLogins:       5,
		SessionTimeout:        30,
		RequireAudit:          true,
		RequireEncryption:     true,
		DataRetention:         2555, // 7 years
		RequireApprovals:      true,
	},
	ModePCIDSS: {
		Mode:                  ModePCIDSS,
		Enabled:               true,
		RequireMFA:            true,
		MinPasswordLength:     12,
		PasswordExpiry:        90,
		RequirePasswordChange: true,
		MaxFailedLogins:       6,
		SessionTimeout:        15,
		RequireAudit:          true,
		RequireEncryption:     true,
		DataRetention:         365,
		RequireApprovals:      false,
	},
	ModeGDPR: {
		Mode:                  ModeGDPR,
		Enabled:               true,
		RequireMFA:            false,
		MinPasswordLength:     8,
		PasswordExpiry:        0, // No expiry requirement
		RequirePasswordChange: false,
		MaxFailedLogins:       10,
		SessionTimeout:        30,
		RequireAudit:          true,
		RequireEncryption:     true,
		DataRetention:         365,
		RequireApprovals:      false,
		AllowedCountries:      []string{}, // EU countries should be specified
	},
	ModeFedRAMP: {
		Mode:                  ModeFedRAMP,
		Enabled:               true,
		RequireMFA:            true,
		MinPasswordLength:     15,
		PasswordExpiry:        60,
		RequirePasswordChange: true,
		MaxFailedLogins:       3,
		SessionTimeout:        10,
		RequireAudit:          true,
		RequireEncryption:     true,
		DataRetention:         2555,
		RequireApprovals:      true,
	},
	ModeISO27001: {
		Mode:                  ModeISO27001,
		Enabled:               true,
		RequireMFA:            true,
		MinPasswordLength:     10,
		PasswordExpiry:        90,
		RequirePasswordChange: true,
		MaxFailedLogins:       5,
		SessionTimeout:        30,
		RequireAudit:          true,
		RequireEncryption:     true,
		DataRetention:         365,
		RequireApprovals:      true,
	},
}

func GetPresetConfig(mode Mode) *ComplianceConfig {
	if config, ok := presetConfigs[mode]; ok {
		copy := *config
		return &copy
	}
	return &ComplianceConfig{
		Mode:              ModeNone,
		Enabled:           false,
		MinPasswordLength: 8,
		SessionTimeout:    1440, // 24 hours
		DataRetention:     90,
	}
}

func (c *ComplianceConfig) Validate() error {
	if c.MinPasswordLength < 8 {
		c.MinPasswordLength = 8
	}
	if c.SessionTimeout < 5 {
		c.SessionTimeout = 5
	}
	if c.DataRetention < 1 {
		c.DataRetention = 1
	}
	return nil
}
