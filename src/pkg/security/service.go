package security

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Service orchestrates security scanning operations
type Service struct {
	factory   *ScannerFactory
	config    *ScanConfig
	repo      Repository
	mu        sync.RWMutex
	scanQueue chan *ScanRequest
}

// ScanRequest represents a security scan request
type ScanRequest struct {
	BuildID int
	Target  string
	Config  *ScanConfig
	Result  chan *CompleteScanResult
}

// CompleteScanResult contains all scan results
type CompleteScanResult struct {
	BuildID            int
	VulnerabilityScans []*ScanResult
	SASTScans          []*ScanResult
	SecretScans        []*ScanResult
	LicenseScans       []*ScanResult
	SBOMScans          []*ScanResult
	Summary            *ScanSummary
	Duration           time.Duration
	Error              error
}

// ScanSummary provides aggregate scan information
type ScanSummary struct {
	TotalVulnerabilities int
	CriticalCount        int
	HighCount            int
	MediumCount          int
	LowCount             int
	InfoCount            int
	SecretsFound         int
	LicenseIssues        int
	CodeIssues           int
	Passed               bool
}

// Repository interface for security data persistence
type Repository interface {
	CreateReport(ctx context.Context, report *SecurityReport) error
	GetReportByBuildID(ctx context.Context, buildID int) ([]*SecurityReport, error)
	GetReportByID(ctx context.Context, id int) (*SecurityReport, error)
	ListReports(ctx context.Context, filters map[string]interface{}) ([]*SecurityReport, error)
}

// NewService creates a new security service
func NewService(config *ScanConfig, repo Repository) *Service {
	s := &Service{
		factory:   NewScannerFactory(),
		config:    config,
		repo:      repo,
		scanQueue: make(chan *ScanRequest, 100),
	}

	// Start scan workers
	for i := 0; i < 5; i++ {
		go s.scanWorker()
	}

	return s
}

// ScanBuild performs all enabled security scans for a build
func (s *Service) ScanBuild(ctx context.Context, buildID int, target string) (*CompleteScanResult, error) {
	start := time.Now()

	result := &CompleteScanResult{
		BuildID: buildID,
		Summary: &ScanSummary{},
	}

	config := s.getEffectiveConfig()

	// Run scans in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := []error{}

	// Vulnerability scanning
	if config.EnableVulnScan {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := s.factory.CreateScanner(ScanTypeVulnerability)
			scanResult, err := scanner.Scan(ctx, target)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("vulnerability scan failed: %w", err))
			} else {
				result.VulnerabilityScans = append(result.VulnerabilityScans, scanResult)
				s.updateSummary(result.Summary, scanResult)
			}
		}()
	}

	// SAST scanning
	if config.EnableSAST {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := s.factory.CreateScanner(ScanTypeSAST)
			scanResult, err := scanner.Scan(ctx, target)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("SAST scan failed: %w", err))
			} else {
				result.SASTScans = append(result.SASTScans, scanResult)
				s.updateSummary(result.Summary, scanResult)
			}
		}()
	}

	// Secret scanning
	if config.EnableSecretScan {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := s.factory.CreateScanner(ScanTypeSecret)
			scanResult, err := scanner.Scan(ctx, target)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("secret scan failed: %w", err))
			} else {
				result.SecretScans = append(result.SecretScans, scanResult)
				s.updateSummary(result.Summary, scanResult)
			}
		}()
	}

	// License scanning
	if config.EnableLicenseScan {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := s.factory.CreateScanner(ScanTypeLicense)
			scanResult, err := scanner.Scan(ctx, target)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("license scan failed: %w", err))
			} else {
				result.LicenseScans = append(result.LicenseScans, scanResult)
				s.updateSummary(result.Summary, scanResult)
			}
		}()
	}

	// SBOM generation
	if config.EnableSBOM {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := s.factory.CreateScanner(ScanTypeSBOM)
			scanResult, err := scanner.Scan(ctx, target)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("SBOM generation failed: %w", err))
			} else {
				result.SBOMScans = append(result.SBOMScans, scanResult)
			}
		}()
	}

	// Wait for all scans to complete
	wg.Wait()

	result.Duration = time.Since(start)

	// Check if build should fail based on findings
	result.Summary.Passed = s.evaluatePass(result.Summary, config)

	// Store scan results in database
	if err := s.storeResults(ctx, buildID, result); err != nil {
		log.Printf("Failed to store security scan results: %v", err)
	}

	if len(errors) > 0 {
		result.Error = fmt.Errorf("scan errors: %v", errors)
	}

	log.Printf("Security scan completed for build %d: %d vulnerabilities, %d secrets, %d code issues in %v",
		buildID, result.Summary.TotalVulnerabilities, result.Summary.SecretsFound,
		result.Summary.CodeIssues, result.Duration)

	return result, nil
}

// ScanBuildAsync queues a build for asynchronous scanning
func (s *Service) ScanBuildAsync(buildID int, target string, config *ScanConfig) {
	req := &ScanRequest{
		BuildID: buildID,
		Target:  target,
		Config:  config,
		Result:  make(chan *CompleteScanResult, 1),
	}

	select {
	case s.scanQueue <- req:
		log.Printf("Build %d queued for security scanning", buildID)
	default:
		log.Printf("Warning: Scan queue full, dropping scan for build %d", buildID)
	}
}

// scanWorker processes scan requests from the queue
func (s *Service) scanWorker() {
	for req := range s.scanQueue {
		ctx := context.Background()
		result, err := s.ScanBuild(ctx, req.BuildID, req.Target)
		if err != nil {
			log.Printf("Scan worker error for build %d: %v", req.BuildID, err)
		}
		req.Result <- result
		close(req.Result)
	}
}

// updateSummary updates the scan summary with results
func (s *Service) updateSummary(summary *ScanSummary, result *ScanResult) {
	// Count vulnerabilities
	for severity, count := range result.Summary {
		switch severity {
		case "critical":
			summary.CriticalCount += count
			summary.TotalVulnerabilities += count
		case "high":
			summary.HighCount += count
			summary.TotalVulnerabilities += count
		case "medium":
			summary.MediumCount += count
			summary.TotalVulnerabilities += count
		case "low":
			summary.LowCount += count
			summary.TotalVulnerabilities += count
		case "info":
			summary.InfoCount += count
			summary.TotalVulnerabilities += count
		}
	}

	// Count secrets
	summary.SecretsFound += len(result.Secrets)

	// Count license issues
	for _, license := range result.Licenses {
		if !license.Compatible {
			summary.LicenseIssues++
		}
	}

	// Count code issues
	summary.CodeIssues += len(result.CodeIssues)
}

// evaluatePass determines if the build passes security checks
func (s *Service) evaluatePass(summary *ScanSummary, config *ScanConfig) bool {
	if config.FailOnCritical && summary.CriticalCount > 0 {
		return false
	}

	if config.FailOnHigh && summary.HighCount > 0 {
		return false
	}

	// Always fail on secrets found
	if summary.SecretsFound > 0 {
		return false
	}

	return true
}

// storeResults stores scan results in the database
func (s *Service) storeResults(ctx context.Context, buildID int, result *CompleteScanResult) error {
	// Store vulnerability reports
	for _, scan := range result.VulnerabilityScans {
		report := &SecurityReport{
			BuildID:       buildID,
			ScanType:      ScanTypeVulnerability,
			Tool:          scan.Tool,
			CriticalCount: result.Summary.CriticalCount,
			HighCount:     result.Summary.HighCount,
			MediumCount:   result.Summary.MediumCount,
			LowCount:      result.Summary.LowCount,
			InfoCount:     result.Summary.InfoCount,
			TotalCount:    result.Summary.TotalVulnerabilities,
			Passed:        result.Summary.Passed,
			Details:       s.convertToDetailsMap(scan),
			RawOutput:     scan.RawOutput,
			CreatedAt:     time.Now(),
		}

		if err := s.repo.CreateReport(ctx, report); err != nil {
			return fmt.Errorf("failed to store vulnerability report: %w", err)
		}
	}

	// Store SAST reports
	for _, scan := range result.SASTScans {
		report := &SecurityReport{
			BuildID:    buildID,
			ScanType:   ScanTypeSAST,
			Tool:       scan.Tool,
			TotalCount: len(scan.CodeIssues),
			Passed:     result.Summary.Passed,
			Details:    s.convertToDetailsMap(scan),
			RawOutput:  scan.RawOutput,
			CreatedAt:  time.Now(),
		}

		if err := s.repo.CreateReport(ctx, report); err != nil {
			return fmt.Errorf("failed to store SAST report: %w", err)
		}
	}

	// Store secret scan reports
	for _, scan := range result.SecretScans {
		report := &SecurityReport{
			BuildID:    buildID,
			ScanType:   ScanTypeSecret,
			Tool:       scan.Tool,
			TotalCount: len(scan.Secrets),
			Passed:     len(scan.Secrets) == 0,
			Details:    s.convertToDetailsMap(scan),
			RawOutput:  scan.RawOutput,
			CreatedAt:  time.Now(),
		}

		if err := s.repo.CreateReport(ctx, report); err != nil {
			return fmt.Errorf("failed to store secret scan report: %w", err)
		}
	}

	// Store license reports
	for _, scan := range result.LicenseScans {
		report := &SecurityReport{
			BuildID:    buildID,
			ScanType:   ScanTypeLicense,
			Tool:       scan.Tool,
			TotalCount: len(scan.Licenses),
			Passed:     result.Summary.LicenseIssues == 0,
			Details:    s.convertToDetailsMap(scan),
			RawOutput:  scan.RawOutput,
			CreatedAt:  time.Now(),
		}

		if err := s.repo.CreateReport(ctx, report); err != nil {
			return fmt.Errorf("failed to store license report: %w", err)
		}
	}

	// Store SBOM
	for _, scan := range result.SBOMScans {
		report := &SecurityReport{
			BuildID:   buildID,
			ScanType:  ScanTypeSBOM,
			Tool:      scan.Tool,
			Passed:    true,
			Details:   s.convertToDetailsMap(scan),
			RawOutput: scan.RawOutput,
			CreatedAt: time.Now(),
		}

		if err := s.repo.CreateReport(ctx, report); err != nil {
			return fmt.Errorf("failed to store SBOM: %w", err)
		}
	}

	return nil
}

// convertToDetailsMap converts scan result to map for JSON storage
func (s *Service) convertToDetailsMap(scan *ScanResult) map[string]interface{} {
	details := make(map[string]interface{})

	if len(scan.Vulnerabilities) > 0 {
		details["vulnerabilities"] = scan.Vulnerabilities
	}

	if len(scan.Secrets) > 0 {
		details["secrets"] = scan.Secrets
	}

	if len(scan.Licenses) > 0 {
		details["licenses"] = scan.Licenses
	}

	if len(scan.CodeIssues) > 0 {
		details["code_issues"] = scan.CodeIssues
	}

	if scan.SBOM != nil {
		details["sbom"] = scan.SBOM
	}

	details["summary"] = scan.Summary
	details["duration"] = scan.Duration.String()

	return details
}

// getEffectiveConfig returns the effective scan configuration
func (s *Service) getEffectiveConfig() *ScanConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config != nil {
		return s.config
	}

	// Default configuration
	return &ScanConfig{
		EnableVulnScan:    true,
		EnableSAST:        true,
		EnableSecretScan:  true,
		EnableLicenseScan: true,
		EnableSBOM:        true,
		FailOnCritical:    true,
		FailOnHigh:        false,
		SBOMFormat:        "spdx",
	}
}

// UpdateConfig updates the scan configuration
func (s *Service) UpdateConfig(config *ScanConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetReportsByBuild retrieves all security reports for a build
func (s *Service) GetReportsByBuild(ctx context.Context, buildID int) ([]*SecurityReport, error) {
	return s.repo.GetReportByBuildID(ctx, buildID)
}

// GetReport retrieves a specific security report
func (s *Service) GetReport(ctx context.Context, reportID int) (*SecurityReport, error) {
	return s.repo.GetReportByID(ctx, reportID)
}
