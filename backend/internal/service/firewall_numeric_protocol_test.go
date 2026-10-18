package service_test

import (
	"fmt"
	"testing"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestParseIptablesNumericProtocols(t *testing.T) {
	for number, expected := range map[string]string{"0": "all", "1": "icmp", "6": "tcp", "17": "udp", "47": "47"} {
		raw := fmt.Sprintf("Chain INPUT (policy ACCEPT 0 packets, 0 bytes)\nnum pkts bytes target prot opt in out source destination\n1 0 0 DROP %s -- * * 0.0.0.0/0 0.0.0.0/0\n", number)
		chains, err := service.ParseIptablesOutput(raw)
		if err != nil || len(chains[domain.ChainInput].Rules) != 1 || chains[domain.ChainInput].Rules[0].Protocol != expected {
			t.Errorf("protocol %s: wanted %s, got %+v, err %v", number, expected, chains[domain.ChainInput], err)
		}
	}
}
