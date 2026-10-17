package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/report"
)

func (s *reportService) loadReportFindings(ctx context.Context, tenant, scanID uuid.UUID) ([]*domain.Finding, error) {
	result := []*domain.Finding{}
	for offset := 0; ; {
		page, total, err := s.findingRepo.List(ctx, tenant, domain.FindingFilter{ScanID: &scanID, Limit: 1000, Offset: offset})
		if err != nil {
			return nil, err
		}
		result = append(result, page...)
		offset += len(page)
		if offset >= total {
			return result, nil
		}
		if len(page) == 0 {
			return nil, fmt.Errorf("paginazione finding incompleta")
		}
	}
}

// Preserve exact-group repository filters; explicitly aggregate the selected root and descendants.
func (s *reportService) loadReportScope(ctx context.Context, tenant, groupID uuid.UUID) ([]*domain.Target, []*domain.ReportItem, error) {
	groups, err := s.targetGroupRepo.List(ctx, tenant)
	if err != nil {
		return nil, nil, err
	}
	ids := []uuid.UUID{groupID}
	seen := map[uuid.UUID]bool{groupID: true}
	for i := 0; i < len(ids); i++ {
		for _, g := range groups {
			if g != nil && g.TenantID == tenant && g.ParentID != nil && *g.ParentID == ids[i] && !seen[g.ID] {
				seen[g.ID] = true
				ids = append(ids, g.ID)
			}
		}
	}
	targets := []*domain.Target{}
	scans := []*domain.ReportItem{}
	for _, id := range ids {
		for offset := 0; ; {
			page, total, e := s.targetRepo.List(ctx, tenant, domain.TargetFilter{GroupID: &id, Limit: 1000, Offset: offset})
			if e != nil {
				return nil, nil, e
			}
			targets = append(targets, page...)
			offset += len(page)
			if offset >= total {
				break
			}
			if len(page) == 0 {
				return nil, nil, fmt.Errorf("paginazione target incompleta")
			}
		}
		for offset := 0; ; {
			page, total, e := s.reportRepo.List(ctx, tenant, domain.ReportFilter{GroupID: &id, Limit: 1000, Offset: offset})
			if e != nil {
				return nil, nil, e
			}
			scans = append(scans, page...)
			offset += len(page)
			if offset >= total {
				break
			}
			if len(page) == 0 {
				return nil, nil, fmt.Errorf("paginazione report incompleta")
			}
		}
	}
	return targets, scans, nil
}

func (s *reportService) departmentInventory(ctx context.Context, tenant uuid.UUID, group *domain.TargetGroup, targets []*domain.Target, scans []*domain.ReportItem, findings []*domain.Finding) ([]report.DepartmentReportSection, map[string]string, bool, error) {
	groups, err := s.targetGroupRepo.List(ctx, tenant)
	if err != nil {
		return nil, nil, false, fmt.Errorf("lettura reparti report: %w", err)
	}
	names := map[uuid.UUID]string{group.ID: group.Name}
	for _, g := range groups {
		if g != nil && g.TenantID == tenant {
			names[g.ID] = g.Name
		}
	}
	// Identity comes from group IDs, not display names (which can equal fallback labels).
	sections := map[string]*report.DepartmentReportSection{}
	values := map[*report.DepartmentReportSection]map[string]bool{}
	sectionFor := func(id *uuid.UUID) *report.DepartmentReportSection {
		key, name := "unassigned", "Senza reparto"
		if id != nil && !(*id == group.ID && group.ParentID == nil) {
			key, name = id.String(), names[*id]
			if name == "" {
				name = "Reparto non disponibile"
			}
			if name == "Senza reparto" || name == "Non attribuito" || name == "Reparto non disponibile" {
				name += " [" + id.String()[:8] + "]"
			}
		}
		if sections[key] == nil {
			sections[key] = &report.DepartmentReportSection{Name: name}
			values[sections[key]] = map[string]bool{}
		}
		return sections[key]
	}
	hasDepartments := false
	for _, g := range groups {
		if g != nil && g.TenantID == tenant && g.ParentID != nil && *g.ParentID == group.ID {
			hasDepartments = true
			sectionFor(&g.ID)
		}
	}
	add := func(section *report.DepartmentReportSection, value string) {
		if value == "" {
			return
		}
		values[section][value] = true
	}
	targetDepartments := map[uuid.UUID]*report.DepartmentReportSection{}
	for _, t := range targets {
		if t == nil {
			continue
		}
		section := sectionFor(t.GroupID)
		targetDepartments[t.ID] = section
		add(section, t.Value)
	}
	scanDepartments := map[string]string{}
	scanSections := map[uuid.UUID]*report.DepartmentReportSection{}
	for _, scan := range scans {
		if scan == nil {
			continue
		}
		section, ok := targetDepartments[scan.TargetID]
		if !ok {
			section = sectionFor(scan.GroupID)
		}
		scanDepartments[scan.ScanID.String()] = section.Name
		scanSections[scan.ScanID] = section
	}
	for _, f := range findings {
		if f == nil {
			continue
		}
		section, ok := scanSections[f.ScanID]
		if !ok {
			if sections["unattributed"] == nil {
				sections["unattributed"] = &report.DepartmentReportSection{Name: "Non attribuito"}
				values[sections["unattributed"]] = map[string]bool{}
			}
			section = sections["unattributed"]
		}
		add(section, f.HostIP)
		section.Findings = append(section.Findings, f)
	}
	result := make([]report.DepartmentReportSection, 0, len(sections))
	for _, section := range sections {
		ips := make([]string, 0, len(values[section]))
		for ip := range values[section] {
			ips = append(ips, ip)
		}
		sort.Strings(ips)
		section.Targets = ips
		result = append(result, *section)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, scanDepartments, group.ParentID == nil && hasDepartments, nil
}
