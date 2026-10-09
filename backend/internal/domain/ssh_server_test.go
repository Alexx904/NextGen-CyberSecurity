package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

func TestSSHAuthType_IsValid(t *testing.T) {
	if !domain.SSHAuthTypeKey.IsValid() {
		t.Errorf("SSHAuthTypeKey deve essere valido")
	}
	if !domain.SSHAuthTypePassword.IsValid() {
		t.Errorf("SSHAuthTypePassword deve essere valido")
	}
	if domain.SSHAuthType("CERTIFICATE").IsValid() {
		t.Errorf("tipo non supportato non deve essere valido")
	}
	if domain.SSHAuthType("").IsValid() {
		t.Errorf("stringa vuota non deve essere valida")
	}
}

func TestSSHServerStatus_IsValid(t *testing.T) {
	validStatuses := []domain.SSHServerStatus{
		domain.SSHServerStatusOnline,
		domain.SSHServerStatusOffline,
		domain.SSHServerStatusUnchecked,
	}

	for _, s := range validStatuses {
		if !s.IsValid() {
			t.Errorf("stato %q deve essere valido", s)
		}
	}

	if domain.SSHServerStatus("UNKNOWN").IsValid() {
		t.Errorf("stato non registrato non deve essere valido")
	}
}

func TestSSHServer_Validate(t *testing.T) {
	validServer := func() domain.SSHServer {
		return domain.SSHServer{
			ID:              uuid.New(),
			TenantID:        uuid.New(),
			Name:            "Bastion Host Prod",
			Host:            "192.168.1.100",
			Port:            22,
			Username:        "admin",
			AuthType:        domain.SSHAuthTypeKey,
			EncryptedSecret: "encrypted-payload-sample",
			Status:          domain.SSHServerStatusUnchecked,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	t.Run("Valid server passes validation", func(t *testing.T) {
		s := validServer()
		if err := s.Validate(); err != nil {
			t.Fatalf("validazione server valida fallita: %v", err)
		}
	})

	t.Run("Default status set to UNCHECKED if empty", func(t *testing.T) {
		s := validServer()
		s.Status = ""
		if err := s.Validate(); err != nil {
			t.Fatalf("validazione fallita con status vuoto: %v", err)
		}
		if s.Status != domain.SSHServerStatusUnchecked {
			t.Errorf("atteso status di default UNCHECKED, ottenuto: %s", s.Status)
		}
	})

	t.Run("TenantID nil returns ErrValidation", func(t *testing.T) {
		s := validServer()
		s.TenantID = uuid.Nil
		if err := s.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation, ottenuto: %v", err)
		}
	})

	t.Run("Empty name returns ErrValidation", func(t *testing.T) {
		s := validServer()
		s.Name = "   "
		if err := s.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation, ottenuto: %v", err)
		}
	})

	t.Run("Empty host returns ErrEmptyHost", func(t *testing.T) {
		s := validServer()
		s.Host = "  "
		if err := s.Validate(); !errors.Is(err, domain.ErrEmptyHost) {
			t.Errorf("atteso ErrEmptyHost, ottenuto: %v", err)
		}
	})

	t.Run("Invalid ports return ErrInvalidSSHPort", func(t *testing.T) {
		invalidPorts := []int{0, -1, 65536, 70000}
		for _, port := range invalidPorts {
			s := validServer()
			s.Port = port
			if err := s.Validate(); !errors.Is(err, domain.ErrInvalidSSHPort) {
				t.Errorf("atteso ErrInvalidSSHPort per porta %d, ottenuto: %v", port, err)
			}
		}
	})

	t.Run("Empty username returns ErrEmptyUsername", func(t *testing.T) {
		s := validServer()
		s.Username = ""
		if err := s.Validate(); !errors.Is(err, domain.ErrEmptyUsername) {
			t.Errorf("atteso ErrEmptyUsername, ottenuto: %v", err)
		}
	})

	t.Run("Invalid auth type returns ErrInvalidSSHAuthType", func(t *testing.T) {
		s := validServer()
		s.AuthType = "OAUTH2"
		if err := s.Validate(); !errors.Is(err, domain.ErrInvalidSSHAuthType) {
			t.Errorf("atteso ErrInvalidSSHAuthType, ottenuto: %v", err)
		}
	})

	t.Run("Empty encrypted secret returns ErrEmptySecret", func(t *testing.T) {
		s := validServer()
		s.EncryptedSecret = "   "
		if err := s.Validate(); !errors.Is(err, domain.ErrEmptySecret) {
			t.Errorf("atteso ErrEmptySecret, ottenuto: %v", err)
		}
	})

	t.Run("Invalid status returns ErrValidation", func(t *testing.T) {
		s := validServer()
		s.Status = "INVALID_STATUS"
		if err := s.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation per status invalido, ottenuto: %v", err)
		}
	})
}

func TestFirewallBackup_Validate(t *testing.T) {
	validBackup := func() domain.FirewallBackup {
		return domain.FirewallBackup{
			ID:               uuid.New(),
			TenantID:         uuid.New(),
			ServerID:         uuid.New(),
			RawIptablesRules: "*filter\n:INPUT ACCEPT [0:0]\nCOMMIT",
			Label:            "Pre-rollout snapshot",
			CreatedAt:        time.Now(),
		}
	}

	t.Run("Valid backup passes validation", func(t *testing.T) {
		b := validBackup()
		if err := b.Validate(); err != nil {
			t.Fatalf("validazione backup valida fallita: %v", err)
		}
	})

	t.Run("Nil TenantID returns ErrValidation", func(t *testing.T) {
		b := validBackup()
		b.TenantID = uuid.Nil
		if err := b.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation, ottenuto: %v", err)
		}
	})

	t.Run("Nil ServerID returns ErrValidation", func(t *testing.T) {
		b := validBackup()
		b.ServerID = uuid.Nil
		if err := b.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation, ottenuto: %v", err)
		}
	})

	t.Run("Empty RawIptablesRules returns ErrValidation", func(t *testing.T) {
		b := validBackup()
		b.RawIptablesRules = "   "
		if err := b.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation, ottenuto: %v", err)
		}
	})

	t.Run("Empty Label returns ErrValidation", func(t *testing.T) {
		b := validBackup()
		b.Label = "   "
		if err := b.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("atteso ErrValidation, ottenuto: %v", err)
		}
	})
}
