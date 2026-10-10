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
		{
			name: "Multi-host consolidated network (9 hosts, 5 High, 7 Medium, 2 Low)",
			stats: domain.ScanSummaryStats{
				TotalHosts:  9,
				UpHosts:     9,
				HighCount:   5, // 75
				MediumCount: 7, // 35
				LowCount:    2, // 2 -> total 112 / sqrt(9) = 37.33 -> 100 - 37 = 63
			},
			expectedScore: 63,
			expectedGrade: "C",
		},
		{
			name: "Single host many Highs without Critical - floor at 25 Grade D",
			stats: domain.ScanSummaryStats{
				TotalHosts:  1,
				UpHosts:     1,
				HighCount:   10, // 150 deduction -> clamped to 25 because CriticalCount is 0
				MediumCount: 5,
			},
			expectedScore: 25,
			expectedGrade: "D",
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
