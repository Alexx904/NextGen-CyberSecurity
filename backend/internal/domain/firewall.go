package domain

import (
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ChainName definisce le catene standard di iptables/nftables
type ChainName string

const (
	ChainInput   ChainName = "INPUT"
	ChainOutput  ChainName = "OUTPUT"
	ChainForward ChainName = "FORWARD"
)

// IsValid verifica se il nome della catena è supportato
func (c ChainName) IsValid() bool {
	switch c {
	case ChainInput, ChainOutput, ChainForward:
		return true
	default:
		return false
	}
}

// RuleAction definisce l'azione (target) eseguita dal firewall
type RuleAction string

const (
	ActionAccept RuleAction = "ACCEPT"
	ActionDrop   RuleAction = "DROP"
	ActionReject RuleAction = "REJECT"
	ActionLog    RuleAction = "LOG"
)

// IsValid verifica se l'azione firewall è supportata
func (a RuleAction) IsValid() bool {
	switch a {
	case ActionAccept, ActionDrop, ActionReject, ActionLog:
		return true
	default:
		return false
	}
}

// FirewallRule modella una singola regola attiva in una catena iptables
type FirewallRule struct {
	LineNumber   int        `json:"line_number"`
	Packets      uint64     `json:"packets"`
	Bytes        uint64     `json:"bytes"`
	Target       RuleAction `json:"target"`
	Protocol     string     `json:"protocol"`
	Opt          string     `json:"opt"`
	InInterface  string     `json:"in_interface"`
	OutInterface string     `json:"out_interface"`
	Source       string     `json:"source"`
	Destination  string     `json:"destination"`
	Options      string     `json:"options"`
}

// FirewallChain modella una catena con la sua policy di default, statistiche e regole ordinate
type FirewallChain struct {
	Name          ChainName      `json:"name"`
	DefaultPolicy RuleAction     `json:"default_policy"`
	Packets       uint64         `json:"packets"`
	Bytes         uint64         `json:"bytes"`
	Rules         []FirewallRule `json:"rules"`
}

// FirewallStatus racchiude lo stato globale del firewall di un host e il tracking del Safety Timer
type FirewallStatus struct {
	ServerID             uuid.UUID                    `json:"server_id"`
	Chains               map[ChainName]*FirewallChain `json:"chains"`
	ActiveSafetyTimer    bool                         `json:"active_safety_timer"`
	SafetyTimerExpiresAt *time.Time                   `json:"safety_timer_expires_at,omitempty"`
	SafetyTimeoutSeconds int                          `json:"safety_timeout_seconds"`
}

// FirewallRuleParams definisce i parametri necessari per aggiungere o modificare una regola
type FirewallRuleParams struct {
	Chain        ChainName  `json:"chain"`
	Action       RuleAction `json:"action"`
	Protocol     string     `json:"protocol"` // all, tcp, udp, icmp
	Source       string     `json:"source"`   // e.g. 192.168.1.0/24 o 0.0.0.0/0
	Destination  string     `json:"destination"`
	Port         string     `json:"port"` // Porta o range, es. 22 o 80:443
	InInterface  string     `json:"in_interface"`
	OutInterface string     `json:"out_interface"`
	Comment      string     `json:"comment"`
	Position     int        `json:"position"` // 1-indexed insert, 0 per append
}

// Validate esegue la validazione dei vincoli di consistenza e sicurezza dei parametri della regola
func (p *FirewallRuleParams) Validate() error {
	if !p.Chain.IsValid() {
		return ErrInvalidChainName
	}
	if !p.Action.IsValid() {
		return ErrInvalidRuleAction
	}

	p.Protocol = strings.ToLower(strings.TrimSpace(p.Protocol))
	if p.Protocol == "" {
		p.Protocol = "all"
	}
	switch p.Protocol {
	case "all", "tcp", "udp", "icmp":
		// protocollo valido
	default:
		return fmt.Errorf("%w: protocollo %q non supportato", ErrValidation, p.Protocol)
	}

	// Validazione porta (consentita solo con protocolli TCP o UDP)
	p.Port = strings.TrimSpace(p.Port)
	if p.Port != "" {
		if p.Protocol != "tcp" && p.Protocol != "udp" {
			return fmt.Errorf("%w: la porta può essere specificata solo per protocolli tcp o udp", ErrValidation)
		}
		if err := validatePortSpec(p.Port); err != nil {
			return err
		}
	}

	if p.Position < 0 {
		return fmt.Errorf("%w: posizione negativa non consentita", ErrValidation)
	}
	for _, address := range []string{p.Source, p.Destination} {
		if address == "" {
			continue
		}
		ip, err := netip.ParseAddr(address)
		if err == nil && ip.Is4() {
			continue
		}
		prefix, err := netip.ParsePrefix(address)
		if err != nil || !prefix.Addr().Is4() {
			return fmt.Errorf("%w: indirizzo IPv4 o CIDR non valido", ErrValidation)
		}
	}
	for _, iface := range []string{p.InInterface, p.OutInterface} {
		if iface == "" || iface == "*" {
			continue
		}
		if len(iface) > 15 || !firewallInterfacePattern.MatchString(iface) {
			return fmt.Errorf("%w: interfaccia non valida", ErrValidation)
		}
	}
	if len(p.Comment) > 255 || !utf8.ValidString(p.Comment) || strings.ContainsFunc(p.Comment, unicode.IsControl) {
		return fmt.Errorf("%w: commento non valido (massimo 255 byte, senza controlli)", ErrValidation)
	}

	return nil
}

var firewallInterfacePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.:-]*\+?$`)

func validatePortSpec(portSpec string) error {
	// Supporto a porta singola "80" o range "80:443"
	if strings.Contains(portSpec, ":") {
		parts := strings.Split(portSpec, ":")
		if len(parts) != 2 {
			return fmt.Errorf("%w: specifica range porta non valida", ErrValidation)
		}
		p1, err1 := strconv.Atoi(parts[0])
		p2, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || p1 <= 0 || p1 > 65535 || p2 <= 0 || p2 > 65535 || p1 > p2 {
			return fmt.Errorf("%w: range porta non valido (range consentito 1-65535)", ErrValidation)
		}
		return nil
	}

	portNum, err := strconv.Atoi(portSpec)
	if err != nil || portNum <= 0 || portNum > 65535 {
		return fmt.Errorf("%w: porta %q non valida (range 1-65535)", ErrValidation, portSpec)
	}
	return nil
}
