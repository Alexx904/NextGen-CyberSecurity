package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

func TestDiscoveryDoesNotReuseIdentityAfterMACChange(t *testing.T) {
	tenant := uuid.New()
	runner := &mockScanRunner{runScanFn: func(context.Context, string, scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
		return &scanner.ScanExecutionResult{RawXML: []byte(`<nmaprun><host><status state="up"/><address addr="192.168.1.10" addrtype="ipv4"/><address addr="00:14:22:AA:BB:CC" addrtype="mac"/><hostnames><hostname name="new-device.local"/></hostnames></host></nmaprun>`)}, nil
	}}
	svc := NewDiscoveryService(runner, nil, nil, nil).(*discoveryService)
	svc.hostCache[fmt.Sprintf("%s:192.168.1.10", tenant)] = cachedHostInfo{MAC: "B8:27:EB:12:34:56", Hostname: "old-device.local", Vendor: "Old Vendor", ObservedAt: time.Now()}
	result, err := svc.DiscoverSubnet(context.Background(), tenant, domain.DiscoveryRequest{Subnet: "192.168.1.10"})
	if err != nil {
		t.Fatal(err)
	}
	d := result.Devices[0]
	if d.Vendor != "Dell" || d.Hostname != "new-device.local" {
		t.Fatalf("stale identity: %+v", d)
	}
	cached := svc.hostCache[fmt.Sprintf("%s:192.168.1.10", tenant)]
	if cached.MAC != d.MAC || cached.Hostname != d.Hostname {
		t.Fatalf("stale cache: %+v", cached)
	}
}

func TestPointToPointAddressesAreDiscoverable(t *testing.T) {
	for _, ip := range []string{"192.168.1.0", "192.168.1.1"} {
		if isNetworkOrBroadcast(ip, "192.168.1.0/31") {
			t.Fatalf("usable /31 address excluded: %s", ip)
		}
	}
	if isNetworkOrBroadcast("192.168.1.1", "192.168.1.1/32") {
		t.Fatal("single host excluded")
	}
}
