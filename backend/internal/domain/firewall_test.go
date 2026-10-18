package domain_test

import (
	"errors"
	"testing"

	"nextgen-cybersec/internal/domain"
)

func TestChainName_IsValid(t *testing.T) {
	if !domain.ChainInput.IsValid() {
		t.Errorf("ChainInput deve essere valido")
	}
	if !domain.ChainOutput.IsValid() {
		t.Errorf("ChainOutput deve essere valido")
	}
	if !domain.ChainForward.IsValid() {
		t.Errorf("ChainForward deve essere valido")
	}
	if domain.ChainName("PREROUTING").IsValid() {
		t.Errorf("catena non standard non deve essere valida")
	}
	if domain.ChainName("").IsValid() {
		t.Errorf("stringa vuota non deve essere valida")
	}
}

func TestRuleAction_IsValid(t *testing.T) {
	actions := []domain.RuleAction{
		domain.ActionAccept,
		domain.ActionDrop,
		domain.ActionReject,
		domain.ActionLog,
	}

	for _, a := range actions {
		if !a.IsValid() {
			t.Errorf("azione %q deve essere valida", a)
		}
	}

	if domain.RuleAction("MASQUERADE").IsValid() {
		t.Errorf("azione non gestita non deve essere valida")
	}
	if domain.RuleAction("").IsValid() {
		t.Errorf("stringa vuota non deve essere valida")
	}
}

func TestFirewallRuleParams_Validate(t *testing.T) {
	baseParams := func() domain.FirewallRuleParams {
		return domain.FirewallRuleParams{
			Chain:       domain.ChainInput,
			Action:      domain.ActionAccept,
			Protocol:    "tcp",
			Source:      "192.168.1.0/24",
			Destination: "0.0.0.0/0",
			Port:        "22",
			Comment:     "Allow SSH from management LAN",
			Position:    0,
		}
	}

	t.Run("Valid standard params", func(t *testing.T) {
		p := baseParams()
		if err := p.Validate(); err != nil {
			t.Fatalf("validazione fallita per parametri validi: %v", err)
		}
	})

	t.Run("Valid with port range", func(t *testing.T) {
		p := baseParams()
		p.Port = "80:443"
		if err := p.Validate(); err != nil {
			t.Fatalf("validazione fallita per port range valido: %v", err)
		}
	})

	t.Run("Valid protocol all without port", func(t *testing.T) {
		p := baseParams()
		p.Protocol = "all"
		p.Port = ""
		if err := p.Validate(); err != nil {
			t.Fatalf("validazione fallita per protocollo all senza porta: %v", err)
		}
	})

	t.Run("Invalid chain name", func(t *testing.T) {
		p := baseParams()
		p.Chain = "INVALID"
		if err := p.Validate(); !errors.Is(err, domain.ErrInvalidChainName) {
			t.Errorf("atteso ErrInvalidChainName, ottenuto: %v", err)
		}
	})

	t.Run("Invalid action", func(t *testing.T) {
		p := baseParams()
		p.Action = "ALLOW"
		if err := p.Validate(); !errors.Is(err, domain.ErrInvalidRuleAction) {
			t.Errorf("atteso ErrInvalidRuleAction, ottenuto: %v", err)
		}
	})

	t.Run("Invalid protocol", func(t *testing.T) {
		p := baseParams()
		p.Protocol = "sctp"
		if err := p.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation per protocollo non supportato, ottenuto: %v", err)
		}
	})

	t.Run("Port specified with non-tcp/udp protocol", func(t *testing.T) {
		p := baseParams()
		p.Protocol = "icmp"
		p.Port = "8"
		if err := p.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation per porta su ICMP, ottenuto: %v", err)
		}
	})

	t.Run("Invalid port numbers", func(t *testing.T) {
		invalidPorts := []string{"0", "-1", "65536", "99999", "abc", "80:80:80", "443:80"}
		for _, port := range invalidPorts {
			p := baseParams()
			p.Port = port
			if err := p.Validate(); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("atteso ErrValidation per porta errata %q, ottenuto: %v", port, err)
			}
		}
	})

	t.Run("Negative position returns error", func(t *testing.T) {
		p := baseParams()
		p.Position = -5
		if err := p.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation per posizione negativa, ottenuto: %v", err)
		}
	})
}

func TestFirewallRuleParams_RejectUnsafeNetworkFields(t *testing.T) {
	for _, field := range []string{"source", "destination", "in", "out", "comment"} {
		t.Run(field, func(t *testing.T) {
			p := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept}
			switch field {
			case "source":
				p.Source = "192.0.2.1; id"
			case "destination":
				p.Destination = "$(id)"
			case "in":
				p.InInterface = "eth0; id"
			case "out":
				p.OutInterface = "eth0\nwhoami"
			case "comment":
				p.Comment = "line\nbreak"
			}
			if !errors.Is(p.Validate(), domain.ErrValidation) {
				t.Fatal("unsafe input accepted")
			}
		})
	}
}

func TestFirewallNetworkFieldBoundaries(t *testing.T) {
	for _, address := range []string{"", "0.0.0.0/0", "192.0.2.1", "192.0.2.0/24"} {
		p := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept, Source: address, Destination: address, InInterface: "eth0.42", OutInterface: "veth+"}
		if err := p.Validate(); err != nil {
			t.Fatalf("valid %q: %v", address, err)
		}
	}
	for _, address := range []string{"::1", "2001:db8::/32", "host.example", "-s", "192.0.2.1/33", "192.0.2.1 --jump DROP"} {
		p := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept, Source: address}
		if !errors.Is(p.Validate(), domain.ErrValidation) {
			t.Fatalf("accepted %q", address)
		}
	}
	for _, iface := range []string{"-eth0", "eth+0", "eth0 eth1", "abcdefghijklmnop", "eth$(id)"} {
		p := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept, InInterface: iface}
		if !errors.Is(p.Validate(), domain.ErrValidation) {
			t.Fatalf("accepted interface %q", iface)
		}
	}
}
