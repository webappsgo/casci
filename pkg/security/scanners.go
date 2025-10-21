package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// Scanner interface for security scanning tools
type Scanner interface {
	Scan(ctx context.Context, target string) (*ScanResult, error)
	Name() string
	Type() ScanType
}

// TrivyScanner implements vulnerability scanning using Trivy
type TrivyScanner struct {
	trivyPath string
}

// NewTrivyScanner creates a new Trivy scanner
func NewTrivyScanner() *TrivyScanner {
	// TODO: Extract embedded binary or use system trivy
	return &TrivyScanner{
		trivyPath: "trivy", // Will use system trivy for now
	}
}

func (t *TrivyScanner) Name() string {
	return "trivy"
}

func (t *TrivyScanner) Type() ScanType {
	return ScanTypeVulnerability
}

func (t *TrivyScanner) Scan(ctx context.Context, target string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Type:    ScanTypeVulnerability,
		Tool:    "trivy",
		Summary: make(map[string]int),
	}

	// Run trivy scan
	cmd := exec.CommandContext(ctx, t.trivyPath,
		"fs",
		"--format", "json",
		"--quiet",
		target,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("trivy scan failed: %v", err)
		result.Duration = time.Since(start)
		return result, nil // Return result even on error
	}

	// Parse trivy JSON output
	var trivyOutput struct {
		Results []struct {
			Target          string `json:"Target"`
			Vulnerabilities []struct {
				VulnerabilityID string  `json:"VulnerabilityID"`
				PkgName         string  `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				FixedVersion    string  `json:"FixedVersion"`
				Severity        string  `json:"Severity"`
				Title           string  `json:"Title"`
				Description     string  `json:"Description"`
				CVSS            map[string]struct {
					V3Score float64 `json:"V3Score"`
				} `json:"CVSS"`
				References []string `json:"References"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(output, &trivyOutput); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to parse trivy output: %v", err)
		result.Duration = time.Since(start)
		return result, nil
	}

	// Convert to our format
	for _, r := range trivyOutput.Results {
		for _, v := range r.Vulnerabilities {
			severity := Severity(v.Severity)

			// Get CVSS score
			var cvss float64
			for _, cvssData := range v.CVSS {
				if cvssData.V3Score > cvss {
					cvss = cvssData.V3Score
				}
			}

			vuln := Vulnerability{
				ID:          v.VulnerabilityID,
				Title:       v.Title,
				Description: v.Description,
				Severity:    severity,
				CVSS:        cvss,
				Package:     v.PkgName,
				Version:     v.InstalledVersion,
				FixedIn:     v.FixedVersion,
				References:  v.References,
			}

			result.Vulnerabilities = append(result.Vulnerabilities, vuln)

			// Update summary
			switch severity {
			case SeverityCritical:
				result.Summary["critical"]++
			case SeverityHigh:
				result.Summary["high"]++
			case SeverityMedium:
				result.Summary["medium"]++
			case SeverityLow:
				result.Summary["low"]++
			default:
				result.Summary["info"]++
			}
		}
	}

	result.Success = true
	result.Duration = time.Since(start)
	result.RawOutput = string(output)

	log.Printf("Trivy scan completed: %d vulnerabilities found in %v",
		len(result.Vulnerabilities), result.Duration)

	return result, nil
}

// GitleaksScanner implements secret detection using Gitleaks
type GitleaksScanner struct {
	gitleaksPath string
}

// NewGitleaksScanner creates a new Gitleaks scanner
func NewGitleaksScanner() *GitleaksScanner {
	return &GitleaksScanner{
		gitleaksPath: "gitleaks", // Will use system gitleaks for now
	}
}

func (g *GitleaksScanner) Name() string {
	return "gitleaks"
}

func (g *GitleaksScanner) Type() ScanType {
	return ScanTypeSecret
}

func (g *GitleaksScanner) Scan(ctx context.Context, target string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Type:    ScanTypeSecret,
		Tool:    "gitleaks",
		Summary: make(map[string]int),
	}

	// Run gitleaks scan
	cmd := exec.CommandContext(ctx, g.gitleaksPath,
		"detect",
		"--source", target,
		"--report-format", "json",
		"--report-path", "/tmp/gitleaks-report.json",
		"--no-git",
	)

	output, err := cmd.CombinedOutput()
	// Gitleaks returns exit code 1 if secrets found, which is expected
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		result.Success = false
		result.Error = fmt.Sprintf("gitleaks scan failed: %v", err)
		result.Duration = time.Since(start)
		return result, nil
	}

	// Parse gitleaks JSON output (would read from /tmp/gitleaks-report.json)
	// For now, stub implementation
	result.Success = true
	result.Duration = time.Since(start)
	result.RawOutput = string(output)

	log.Printf("Gitleaks scan completed: %d secrets found in %v",
		len(result.Secrets), result.Duration)

	return result, nil
}

// SemgrepScanner implements SAST using Semgrep
type SemgrepScanner struct {
	semgrepPath string
}

// NewSemgrepScanner creates a new Semgrep scanner
func NewSemgrepScanner() *SemgrepScanner {
	return &SemgrepScanner{
		semgrepPath: "semgrep", // Will use system semgrep for now
	}
}

func (s *SemgrepScanner) Name() string {
	return "semgrep"
}

func (s *SemgrepScanner) Type() ScanType {
	return ScanTypeSAST
}

func (s *SemgrepScanner) Scan(ctx context.Context, target string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Type:    ScanTypeSAST,
		Tool:    "semgrep",
		Summary: make(map[string]int),
	}

	// Run semgrep scan
	cmd := exec.CommandContext(ctx, s.semgrepPath,
		"--config", "auto",
		"--json",
		"--quiet",
		target,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("semgrep scan failed: %v", err)
		result.Duration = time.Since(start)
		return result, nil
	}

	// Parse semgrep JSON output (stub for now)
	result.Success = true
	result.Duration = time.Since(start)
	result.RawOutput = string(output)

	log.Printf("Semgrep scan completed: %d issues found in %v",
		len(result.CodeIssues), result.Duration)

	return result, nil
}

// SyftScanner implements SBOM generation using Syft
type SyftScanner struct {
	syftPath string
}

// NewSyftScanner creates a new Syft scanner
func NewSyftScanner() *SyftScanner {
	return &SyftScanner{
		syftPath: "syft", // Will use system syft for now
	}
}

func (s *SyftScanner) Name() string {
	return "syft"
}

func (s *SyftScanner) Type() ScanType {
	return ScanTypeSBOM
}

func (s *SyftScanner) Scan(ctx context.Context, target string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Type:    ScanTypeSBOM,
		Tool:    "syft",
		Summary: make(map[string]int),
	}

	// Run syft scan
	cmd := exec.CommandContext(ctx, s.syftPath,
		target,
		"-o", "spdx-json",
		"--quiet",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("syft scan failed: %v", err)
		result.Duration = time.Since(start)
		return result, nil
	}

	// Parse syft SPDX output (stub for now)
	result.Success = true
	result.Duration = time.Since(start)
	result.RawOutput = string(output)

	log.Printf("Syft SBOM generated in %v", result.Duration)

	return result, nil
}

// GrypeScanner implements vulnerability matching using Grype
type GrypeScanner struct {
	grypePath string
}

// NewGrypeScanner creates a new Grype scanner
func NewGrypeScanner() *GrypeScanner {
	return &GrypeScanner{
		grypePath: "grype", // Will use system grype for now
	}
}

func (g *GrypeScanner) Name() string {
	return "grype"
}

func (g *GrypeScanner) Type() ScanType {
	return ScanTypeVulnerability
}

func (g *GrypeScanner) Scan(ctx context.Context, target string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Type:    ScanTypeVulnerability,
		Tool:    "grype",
		Summary: make(map[string]int),
	}

	// Run grype scan
	cmd := exec.CommandContext(ctx, g.grypePath,
		"dir:"+target,
		"-o", "json",
		"--quiet",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("grype scan failed: %v", err)
		result.Duration = time.Since(start)
		return result, nil
	}

	// Parse grype JSON output (stub for now)
	result.Success = true
	result.Duration = time.Since(start)
	result.RawOutput = string(output)

	log.Printf("Grype scan completed: %d vulnerabilities found in %v",
		len(result.Vulnerabilities), result.Duration)

	return result, nil
}

// LicenseScannerImpl implements license scanning
type LicenseScannerImpl struct {
	// Use multiple tools or custom implementation
}

// NewLicenseScanner creates a new license scanner
func NewLicenseScanner() *LicenseScannerImpl {
	return &LicenseScannerImpl{}
}

func (l *LicenseScannerImpl) Name() string {
	return "license-scanner"
}

func (l *LicenseScannerImpl) Type() ScanType {
	return ScanTypeLicense
}

func (l *LicenseScannerImpl) Scan(ctx context.Context, target string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Type:    ScanTypeLicense,
		Tool:    "license-scanner",
		Summary: make(map[string]int),
	}

	// Stub implementation
	// TODO: Implement license scanning logic
	result.Success = true
	result.Duration = time.Since(start)

	log.Printf("License scan completed in %v", result.Duration)

	return result, nil
}

// ScannerFactory creates scanners based on type
type ScannerFactory struct{}

// NewScannerFactory creates a new scanner factory
func NewScannerFactory() *ScannerFactory {
	return &ScannerFactory{}
}

// CreateScanner creates a scanner for the given type
func (f *ScannerFactory) CreateScanner(scanType ScanType) Scanner {
	switch scanType {
	case ScanTypeVulnerability:
		return NewTrivyScanner() // Or Grype
	case ScanTypeSAST:
		return NewSemgrepScanner()
	case ScanTypeSecret:
		return NewGitleaksScanner()
	case ScanTypeSBOM:
		return NewSyftScanner()
	case ScanTypeLicense:
		return NewLicenseScanner()
	default:
		return nil
	}
}

// CreateAllScanners creates all available scanners
func (f *ScannerFactory) CreateAllScanners() []Scanner {
	return []Scanner{
		NewTrivyScanner(),
		NewSemgrepScanner(),
		NewGitleaksScanner(),
		NewSyftScanner(),
		NewLicenseScanner(),
	}
}
