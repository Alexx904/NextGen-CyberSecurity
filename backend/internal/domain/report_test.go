package domain_test

import (
	"testing"

	"nextgen-cybersec/internal/domain"
)

func TestReportTemplate_IsValid(t *testing.T) {
	tests := []struct {
		tmpl     domain.ReportTemplate
		expected bool
	}{
		{domain.ReportTemplateExecutive, true},
		{domain.ReportTemplateTechnical, true},
		{domain.ReportTemplate("invalid"), false},
		{domain.ReportTemplate(""), false},
		{domain.ReportTemplate("EXECUTIVE"), false},
	}

	for _, tt := range tests {
		if got := tt.tmpl.IsValid(); got != tt.expected {
			t.Errorf("ReportTemplate(%q).IsValid() = %v, atteso %v", tt.tmpl, got, tt.expected)
		}
	}
}

func TestCalculatePostureScore(t *testing.T) {
	tests := []struct {
		name          string
		stats         domain.ScanSummaryStats
		expectedScore int
		expectedGrade string
	}{
		{
			name:          "Zero vulnerabilities - Pristine posture",
			stats:         domain.ScanSummaryStats{},
			expectedScore: 100,
			expectedGrade: "A",
		},
		{
			name: "Low vulnerabilities only",
			stats: domain.ScanSummaryStats{
				LowCount: 5,
			},
			expectedScore: 95,
			expectedGrade: "A",
		},
		{
			name: "Medium and low vulnerabilities",
			stats: domain.ScanSummaryStats{
				MediumCount: 3, // -15
				LowCount:    5, // -5
			},
			expectedScore: 80,
			expectedGrade: "B",
		},
		{
			name: "Single High vulnerability",
			stats: domain.ScanSummaryStats{
				HighCount: 2, // -30
			},
			expectedScore: 70,
			expectedGrade: "C",
		},
		{
			name: "Multiple High and Medium",
			stats: domain.ScanSummaryStats{
				HighCount:   3, // -45
				MediumCount: 4, // -20
			},
			expectedScore: 35,
			expectedGrade: "D",
		},
		{
			name: "Critical vulnerability present",
			stats: domain.ScanSummaryStats{
				CriticalCount: 2, // -50
				HighCount:     3, // -45
			},
			expectedScore: 5,
			expectedGrade: "F",
		},
		{
			name: "Heavy vulnerabilities clamping at 0",
			stats: domain.ScanSummaryStats{
				CriticalCount: 10, // -250
			},
			expectedScore: 0,
			expectedGrade: "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, grade := domain.CalculatePostureScore(tt.stats)
			if score != tt.expectedScore {
				t.Errorf("CalculatePostureScore() score = %d, atteso %d", score, tt.expectedScore)
			}
			if grade != tt.expectedGrade {
				t.Errorf("CalculatePostureScore() grade = %s, atteso %s", grade, tt.expectedGrade)
			}
		})
	}
}

func TestCalculateMaxSeverity(t *testing.T) {
	tests := []struct {
		name     string
		stats    domain.ScanSummaryStats
		expected domain.SeverityLevel
	}{
		{
			name: "Critical present",
			stats: domain.ScanSummaryStats{
				CriticalCount: 1,
				HighCount:     2,
			},
			expected: domain.SeverityCritical,
		},
		{
			name: "High present (no critical)",
			stats: domain.ScanSummaryStats{
				HighCount:   3,
				MediumCount: 5,
			},
			expected: domain.SeverityHigh,
		},
		{
			name: "Medium present (no high or critical)",
			stats: domain.ScanSummaryStats{
				MediumCount: 2,
				LowCount:    1,
			},
			expected: domain.SeverityMedium,
		},
		{
			name: "Low present",
			stats: domain.ScanSummaryStats{
				LowCount: 4,
			},
			expected: domain.SeverityLow,
		},
		{
			name: "Info only or clean",
			stats: domain.ScanSummaryStats{
				InfoCount: 2,
			},
			expected: domain.SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.CalculateMaxSeverity(tt.stats); got != tt.expected {
				t.Errorf("CalculateMaxSeverity() = %v, atteso %v", got, tt.expected)
			}
		})
	}
}
