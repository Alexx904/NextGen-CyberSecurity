package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

func authorizedScanContext(tenant uuid.UUID, profile *domain.ScanProfile, targets ...*domain.Target) context.Context {
	request := &domain.ScanAuthorizationRequest{Accepted: true, Statement: domain.ScanAuthorizationStatement, NmapArgs: profile.NmapArgs, TimeoutSeconds: profile.TimeoutSeconds}
	for _, target := range targets {
		request.Targets = append(request.Targets, domain.AuthorizedTarget{ID: target.ID, Value: target.Value})
	}
	return domain.WithScanAuthorization(context.Background(), tenant, uuid.New(), request)
}

func TestSchedulerNextRunMatchesServerClock(t *testing.T) {
	previous := time.Local
	time.Local = time.FixedZone("QA server", 2*60*60)
	t.Cleanup(func() { time.Local = previous })
	scheduler := NewSchedulerService(newMockScheduleRepo(), newMockScanRepo(), &mockOrchestratorClient{}, nil)
	schedule := &domain.Schedule{ID: uuid.New(), TenantID: uuid.New(), TargetID: uuid.New(), ProfileID: uuid.New(), CronExpression: "0 12 * * *", IsActive: false}
	ctx := authorizeScheduleFixture(scheduler, schedule)
	if err := scheduler.AddSchedule(ctx, schedule); err != nil {
		t.Fatal(err)
	}
	if schedule.NextRunAt == nil || schedule.NextRunAt.In(time.Local).Hour() != 12 {
		t.Fatalf("next run disagrees with authorized server-local cron: %v", schedule.NextRunAt)
	}
}

func TestWorkerAuthorizationRejectsChangedConfigurationBeforeNmap(t *testing.T) {
	for _, mutation := range []string{"missing", "target", "options", "timeout", "tenant", "disabled schedule"} {
		t.Run(mutation, func(t *testing.T) {
			tenant, targetID, profileID := uuid.New(), uuid.New(), uuid.New()
			target := &domain.Target{ID: targetID, TenantID: tenant, Value: "192.168.1.20"}
			profile := &domain.ScanProfile{ID: profileID, NmapArgs: "-sS", TimeoutSeconds: 60}
			targets := &mockTargetRepo{targets: map[uuid.UUID]*domain.Target{targetID: target}}
			profiles := &mockProfileRepo{profiles: map[uuid.UUID]*domain.ScanProfile{profileID: profile}}
			scan := &domain.Scan{ID: uuid.New(), TenantID: tenant, TargetID: targetID, ProfileID: profileID, TriggerType: domain.TriggerManual, Status: domain.ScanStatusQueued}
			authorizeScanFixture(scan, targets, profiles)
			schedules := newMockScheduleRepo()
			switch mutation {
			case "missing":
				scan.Authorization = nil
			case "target":
				target.Value = "192.168.1.21"
			case "options":
				profile.NmapArgs = "-sS -T5"
			case "timeout":
				profile.TimeoutSeconds++
			case "tenant":
				other := uuid.New()
				profile.TenantID = &other
			case "disabled schedule":
				id := uuid.New()
				scan.TriggerType = domain.TriggerScheduled
				scan.Authorization.ScheduleID = &id
				schedules.schedules[id] = &domain.Schedule{ID: id, TenantID: tenant, TargetID: targetID, ProfileID: profileID, Authorization: scan.Authorization, IsActive: false}
			}
			scans := newMockScanRepo()
			_ = scans.Create(context.Background(), scan)
			called := false
			runner := &mockRunner{runFunc: func(context.Context, string, scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
				called = true
				return nil, nil
			}}
			o := NewScanOrchestrator(OrchestratorConfig{MaxWorkers: 1, QueueSize: 1}, scans, profiles, targets, runner, silentLogger())
			o.SetScheduleRepo(schedules)
			o.ctx = context.Background()
			o.processJob(&ScanJob{TenantID: tenant, ScanID: scan.ID}, 0)
			result, _ := scans.GetByID(context.Background(), tenant, scan.ID)
			if called || result.Status != domain.ScanStatusFailed {
				t.Fatalf("unauthorized worker execution: called=%v status=%v", called, result.Status)
			}
		})
	}
}

func authorizeScanFixture(scan *domain.Scan, targets *mockTargetRepo, profiles *mockProfileRepo) {
	target, profile := targets.targets[scan.TargetID], profiles.profiles[scan.ProfileID]
	if target != nil && profile != nil {
		scan.Authorization = domain.ScanAuthorizationFromContext(authorizedScanContext(scan.TenantID, profile, target))
	}
}

func authorizeScheduleFixture(scheduler *SchedulerService, schedule *domain.Schedule) context.Context {
	target := &domain.Target{ID: schedule.TargetID, TenantID: schedule.TenantID, Value: "192.168.1.20"}
	profile := &domain.ScanProfile{ID: schedule.ProfileID, NmapArgs: "-sS -T3", TimeoutSeconds: 60}
	scheduler.ConfigureAuthorizationRepositories(&mockTargetRepo{targets: map[uuid.UUID]*domain.Target{target.ID: target}}, &mockProfileRepo{profiles: map[uuid.UUID]*domain.ScanProfile{profile.ID: profile}})
	ctx := authorizedScanContext(schedule.TenantID, profile, target)
	auth := domain.ScanAuthorizationFromContext(ctx)
	auth.CronExpression = schedule.CronExpression
	auth.ScheduleID = &schedule.ID
	schedule.Authorization = auth
	return ctx
}

func TestScanAuthorizationMissingAndStaleDoNotEnqueue(t *testing.T) {
	tenant, targetID, profileID := uuid.New(), uuid.New(), uuid.New()
	target := &domain.Target{ID: targetID, TenantID: tenant, Value: "192.168.1.20"}
	profile := &domain.ScanProfile{ID: profileID, NmapArgs: "-sS", TimeoutSeconds: 60}
	scans, queue := newFullMockScanRepo(), &mockOrchestratorClient{}
	targets := &mockTargetRepo{targets: map[uuid.UUID]*domain.Target{targetID: target}}
	profiles := &mockProfileRepo{profiles: map[uuid.UUID]*domain.ScanProfile{profileID: profile}}
	svc := NewScanService(scans, profiles, targets, newMockFindingRepo(), newMockScanGroupRepo(), queue)
	for _, ctx := range []context.Context{context.Background(), authorizedScanContext(tenant, profile, &domain.Target{ID: targetID, TenantID: tenant, Value: "192.168.1.21"})} {
		if _, err := svc.StartScan(ctx, tenant, targetID, profileID); !errors.Is(err, domain.ErrScanAuthorizationRequired) {
			t.Fatalf("got %v", err)
		}
	}
	if len(scans.scans) != 0 || len(queue.enqueuedScans) != 0 {
		t.Fatal("unauthorized work created")
	}
	if _, err := svc.StartScan(authorizedScanContext(tenant, profile, target), tenant, targetID, profileID); err != nil {
		t.Fatal(err)
	}
	if len(queue.enqueuedScans) != 1 {
		t.Fatal("authorized scan not queued")
	}
}

func TestSchedulerAuthorizationRejectsLegacyAndStaleCallback(t *testing.T) {
	repo, scans, queue := newMockScheduleRepo(), newMockScanRepo(), &mockOrchestratorClient{}
	scheduler := NewSchedulerService(repo, scans, queue, nil)
	schedule := &domain.Schedule{ID: uuid.New(), TenantID: uuid.New(), TargetID: uuid.New(), ProfileID: uuid.New(), CronExpression: "0 2 * * *", IsActive: true}
	ctx := authorizeScheduleFixture(scheduler, schedule)
	schedule.Authorization = nil
	_ = repo.Create(ctx, schedule)
	scheduler.TriggerJob(schedule)
	if len(queue.enqueuedScans) != 0 {
		t.Fatal("legacy schedule ran")
	}
	schedule.Authorization = domain.ScanAuthorizationFromContext(ctx)
	_ = repo.Update(ctx, schedule)
	stale := *schedule
	current := *schedule
	current.IsActive = false
	_ = repo.Update(ctx, &current)
	scheduler.TriggerJob(&stale)
	if len(queue.enqueuedScans) != 0 {
		t.Fatal("disabled schedule ran from stale callback")
	}
	_ = repo.Delete(ctx, schedule.TenantID, schedule.ID)
	scheduler.TriggerJob(&stale)
	if len(queue.enqueuedScans) != 0 {
		t.Fatal("deleted schedule ran")
	}
}
