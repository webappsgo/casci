package compliance

import "time"

type Mode string

const (
	ModeNone    Mode = "none"
	ModeHIPAA   Mode = "hipaa"
	ModeSOX     Mode = "sox"
	ModePCIDSS  Mode = "pci-dss"
	ModeGDPR    Mode = "gdpr"
	ModeFedRAMP Mode = "fedramp"
	ModeISO27001 Mode = "iso27001"
)

type ComplianceConfig struct {
	Mode                  Mode   `json:"mode"`
	Enabled               bool   `json:"enabled"`
	RequireMFA            bool   `json:"require_mfa"`
	MinPasswordLength     int    `json:"min_password_length"`
	PasswordExpiry        int    `json:"password_expiry_days"`
	RequirePasswordChange bool   `json:"require_password_change"`
	MaxFailedLogins       int    `json:"max_failed_logins"`
	SessionTimeout        int    `json:"session_timeout_minutes"`
	RequireAudit          bool   `json:"require_audit"`
	RequireEncryption     bool   `json:"require_encryption"`
	DataRetention         int    `json:"data_retention_days"`
	RequireApprovals      bool   `json:"require_approvals"`
	AllowedCountries      []string `json:"allowed_countries,omitempty"`
	DeniedCountries       []string `json:"denied_countries,omitempty"`
}

type ComplianceReport struct {
	ID             int64     `json:"id"`
	GeneratedAt    time.Time `json:"generated_at"`
	Mode           Mode      `json:"mode"`
	Status         string    `json:"status"`
	Compliant      bool      `json:"compliant"`
	TotalChecks    int       `json:"total_checks"`
	PassedChecks   int       `json:"passed_checks"`
	FailedChecks   int       `json:"failed_checks"`
	SkippedChecks  int       `json:"skipped_checks"`
	Findings       []Finding `json:"findings"`
	Recommendations []string  `json:"recommendations"`
}

type Finding struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	CheckID     string `json:"check_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Details     string `json:"details,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

type ComplianceCheck struct {
	ID          string
	Name        string
	Description string
	Category    string
	Severity    string
	Mode        []Mode
	CheckFunc   func() (bool, string, error)
}

const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
	
	StatusPassed     = "passed"
	StatusFailed     = "failed"
	StatusSkipped    = "skipped"
	StatusError      = "error"
	
	CategoryAccess      = "access_control"
	CategoryEncryption  = "encryption"
	CategoryAudit       = "audit"
	CategoryData        = "data_protection"
	CategoryNetwork     = "network_security"
	CategoryBackup      = "backup_recovery"
	CategoryIncident    = "incident_response"
	CategoryCompliance  = "compliance"
)
