package security

import (
	"time"
)

// ScanType represents the type of security scan
type ScanType string

const (
	ScanTypeVulnerability ScanType = "vulnerability"
	ScanTypeSAST          ScanType = "sast"
	ScanTypeSecret        ScanType = "secret"
	ScanTypeLicense       ScanType = "license"
	ScanTypeSBOM          ScanType = "sbom"
	ScanTypeMalware       ScanType = "malware"
)

// Severity represents vulnerability severity
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// SecurityReport represents a complete security scan report
type SecurityReport struct {
	ID            int                    `json:"id"`
	BuildID       int                    `json:"build_id"`
	ScanType      ScanType               `json:"scan_type"`
	Tool          string                 `json:"tool"` // trivy, semgrep, gitleaks, etc.
	CriticalCount int                    `json:"critical_count"`
	HighCount     int                    `json:"high_count"`
	MediumCount   int                    `json:"medium_count"`
	LowCount      int                    `json:"low_count"`
	InfoCount     int                    `json:"info_count"`
	TotalCount    int                    `json:"total_count"`
	Passed        bool                   `json:"passed"`
	Details       map[string]interface{} `json:"details"`
	RawOutput     string                 `json:"raw_output,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// Vulnerability represents a single vulnerability
type Vulnerability struct {
	ID          string   `json:"id"` // CVE-2024-1234
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	CVSS        float64  `json:"cvss,omitempty"`
	Package     string   `json:"package"`
	Version     string   `json:"version"`
	FixedIn     string   `json:"fixed_in,omitempty"`
	References  []string `json:"references,omitempty"`
}

// Secret represents a detected secret
type Secret struct {
	Type        string `json:"type"` // AWS Key, GitHub Token, etc.
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Match       string `json:"match,omitempty"`
	Entropy     float64 `json:"entropy,omitempty"`
}

// License represents a license finding
type License struct {
	Name       string `json:"name"` // MIT, Apache-2.0, GPL-3.0
	Package    string `json:"package"`
	Version    string `json:"version"`
	Compatible bool   `json:"compatible"`
	Risk       string `json:"risk"` // high, medium, low
}

// CodeIssue represents a SAST finding
type CodeIssue struct {
	RuleID      string   `json:"rule_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Column      int      `json:"column,omitempty"`
	Code        string   `json:"code,omitempty"`
	Fix         string   `json:"fix,omitempty"`
	References  []string `json:"references,omitempty"`
}

// SBOM represents a Software Bill of Materials
type SBOM struct {
	Format       string      `json:"format"` // SPDX, CycloneDX
	Version      string      `json:"version"`
	Components   []Component `json:"components"`
	Dependencies []Dependency `json:"dependencies"`
	GeneratedAt  time.Time   `json:"generated_at"`
}

// Component represents a software component
type Component struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Type      string            `json:"type"` // library, application, framework
	Licenses  []string          `json:"licenses,omitempty"`
	CPE       string            `json:"cpe,omitempty"`
	PURL      string            `json:"purl,omitempty"`
	Checksums map[string]string `json:"checksums,omitempty"`
}

// Dependency represents a dependency relationship
type Dependency struct {
	Ref          string   `json:"ref"`
	DependsOn    []string `json:"depends_on,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"` // For SPDX format
}

// ScanConfig represents security scan configuration
type ScanConfig struct {
	EnableVulnScan   bool     `json:"enable_vuln_scan"`
	EnableSAST       bool     `json:"enable_sast"`
	EnableSecretScan bool     `json:"enable_secret_scan"`
	EnableLicenseScan bool    `json:"enable_license_scan"`
	EnableSBOM       bool     `json:"enable_sbom"`
	FailOnCritical   bool     `json:"fail_on_critical"`
	FailOnHigh       bool     `json:"fail_on_high"`
	IgnoreCVEs       []string `json:"ignore_cves,omitempty"`
	AllowedLicenses  []string `json:"allowed_licenses,omitempty"`
	SBOMFormat       string   `json:"sbom_format"` // spdx, cyclonedx
}

// ScanResult represents the result of a scan
type ScanResult struct {
	Type           ScanType               `json:"type"`
	Tool           string                 `json:"tool"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
	Vulnerabilities []Vulnerability       `json:"vulnerabilities,omitempty"`
	Secrets        []Secret               `json:"secrets,omitempty"`
	Licenses       []License              `json:"licenses,omitempty"`
	CodeIssues     []CodeIssue            `json:"code_issues,omitempty"`
	SBOM           *SBOM                  `json:"sbom,omitempty"`
	Summary        map[string]int         `json:"summary"`
	RawOutput      string                 `json:"raw_output,omitempty"`
	Duration       time.Duration          `json:"duration"`
}

// PolicyViolation represents a security policy violation
type PolicyViolation struct {
	Type        string   `json:"type"` // vulnerability, secret, license
	Severity    Severity `json:"severity"`
	Description string   `json:"description"`
	Resource    string   `json:"resource"`
	Action      string   `json:"action"` // block, warn, ignore
}
