package compliance

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Service struct {
	config *ComplianceConfig
	checks []ComplianceCheck
}

func NewService(config *ComplianceConfig) *Service {
	if config == nil {
		config = GetPresetConfig(ModeNone)
	}
	
	s := &Service{
		config: config,
		checks: make([]ComplianceCheck, 0),
	}
	
	s.registerChecks()
	
	return s
}

func (s *Service) GetConfig() *ComplianceConfig {
	return s.config
}

func (s *Service) SetConfig(config *ComplianceConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	s.config = config
	return nil
}

func (s *Service) SetMode(mode Mode) error {
	config := GetPresetConfig(mode)
	return s.SetConfig(config)
}

func (s *Service) RunCompliance(ctx context.Context) (*ComplianceReport, error) {
	report := &ComplianceReport{
		GeneratedAt: time.Now(),
		Mode:        s.config.Mode,
		Status:      "running",
		Findings:    make([]Finding, 0),
		Recommendations: make([]string, 0),
	}
	
	if !s.config.Enabled {
		report.Status = "disabled"
		report.Compliant = true
		return report, nil
	}
	
	for _, check := range s.checks {
		if !s.shouldRunCheck(check) {
			report.SkippedChecks++
			continue
		}
		
		report.TotalChecks++
		
		passed, details, err := check.CheckFunc()
		
		finding := Finding{
			Severity:    check.Severity,
			Category:    check.Category,
			CheckID:     check.ID,
			Description: check.Description,
			Details:     details,
		}
		
		if err != nil {
			finding.Status = StatusError
			finding.Details = fmt.Sprintf("Error: %v. Details: %s", err, details)
			report.FailedChecks++
			report.Findings = append(report.Findings, finding)
			continue
		}
		
		if passed {
			finding.Status = StatusPassed
			report.PassedChecks++
		} else {
			finding.Status = StatusFailed
			report.FailedChecks++
			report.Findings = append(report.Findings, finding)
		}
	}
	
	report.Compliant = report.FailedChecks == 0
	report.Status = "completed"
	
	if !report.Compliant {
		report.Recommendations = s.generateRecommendations(report)
	}
	
	log.Printf("Compliance check completed: %d/%d checks passed", report.PassedChecks, report.TotalChecks)
	
	return report, nil
}

func (s *Service) shouldRunCheck(check ComplianceCheck) bool {
	if len(check.Mode) == 0 {
		return true
	}
	
	for _, mode := range check.Mode {
		if mode == s.config.Mode {
			return true
		}
	}
	
	return false
}

func (s *Service) generateRecommendations(report *ComplianceReport) []string {
	recommendations := make([]string, 0)
	
	criticalFailed := 0
	highFailed := 0
	
	for _, finding := range report.Findings {
		if finding.Status == StatusFailed {
			if finding.Severity == SeverityCritical {
				criticalFailed++
			} else if finding.Severity == SeverityHigh {
				highFailed++
			}
		}
	}
	
	if criticalFailed > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d critical findings immediately", criticalFailed))
	}
	
	if highFailed > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Review and remediate %d high-severity findings", highFailed))
	}
	
	if s.config.Mode != ModeNone {
		recommendations = append(recommendations,
			fmt.Sprintf("Ensure all requirements for %s compliance are met", s.config.Mode))
	}
	
	return recommendations
}

func (s *Service) registerChecks() {
	s.checks = []ComplianceCheck{
		{
			ID:          "AUDIT-001",
			Name:        "Audit Logging Enabled",
			Description: "Verify that audit logging is enabled",
			Category:    CategoryAudit,
			Severity:    SeverityCritical,
			Mode:        []Mode{ModeHIPAA, ModeSOX, ModePCIDSS, ModeFedRAMP, ModeISO27001},
			CheckFunc: func() (bool, string, error) {
				return s.config.RequireAudit, "Audit logging configuration", nil
			},
		},
		{
			ID:          "ENC-001",
			Name:        "Encryption Required",
			Description: "Verify that encryption is enforced",
			Category:    CategoryEncryption,
			Severity:    SeverityCritical,
			Mode:        []Mode{ModeHIPAA, ModeSOX, ModePCIDSS, ModeGDPR, ModeFedRAMP, ModeISO27001},
			CheckFunc: func() (bool, string, error) {
				return s.config.RequireEncryption, "Encryption configuration", nil
			},
		},
		{
			ID:          "AUTH-001",
			Name:        "Password Length",
			Description: "Verify minimum password length meets requirements",
			Category:    CategoryAccess,
			Severity:    SeverityHigh,
			Mode:        []Mode{},
			CheckFunc: func() (bool, string, error) {
				passed := s.config.MinPasswordLength >= 8
				details := fmt.Sprintf("Minimum password length: %d", s.config.MinPasswordLength)
				return passed, details, nil
			},
		},
		{
			ID:          "AUTH-002",
			Name:        "MFA Enabled",
			Description: "Verify that multi-factor authentication is required",
			Category:    CategoryAccess,
			Severity:    SeverityCritical,
			Mode:        []Mode{ModeHIPAA, ModeSOX, ModePCIDSS, ModeFedRAMP, ModeISO27001},
			CheckFunc: func() (bool, string, error) {
				return s.config.RequireMFA, "MFA configuration", nil
			},
		},
		{
			ID:          "AUTH-003",
			Name:        "Session Timeout",
			Description: "Verify session timeout is appropriately configured",
			Category:    CategoryAccess,
			Severity:    SeverityMedium,
			Mode:        []Mode{},
			CheckFunc: func() (bool, string, error) {
				passed := s.config.SessionTimeout <= 60
				details := fmt.Sprintf("Session timeout: %d minutes", s.config.SessionTimeout)
				return passed, details, nil
			},
		},
		{
			ID:          "DATA-001",
			Name:        "Data Retention Policy",
			Description: "Verify data retention policy is configured",
			Category:    CategoryData,
			Severity:    SeverityMedium,
			Mode:        []Mode{},
			CheckFunc: func() (bool, string, error) {
				passed := s.config.DataRetention > 0
				details := fmt.Sprintf("Data retention: %d days", s.config.DataRetention)
				return passed, details, nil
			},
		},
	}
}

func (s *Service) ValidatePassword(password string) error {
	if len(password) < s.config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", s.config.MinPasswordLength)
	}
	return nil
}

func (s *Service) IsCompliant(check string) bool {
	for _, c := range s.checks {
		if c.ID == check {
			passed, _, _ := c.CheckFunc()
			return passed
		}
	}
	return false
}
