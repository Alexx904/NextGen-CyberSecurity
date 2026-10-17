package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/crypto"
)

// SSHConnectionTestResult modella il risultato del test di connettività verso un server remoto
type SSHConnectionTestResult struct {
	Online    bool   `json:"online"`
	LatencyMs int64  `json:"latency_ms"`
	OSInfo    string `json:"os_info"`
	IsRoot    bool   `json:"is_root"`
	Error     string `json:"error,omitempty"`
}

// CreateSSHServerRequest racchiude i parametri per registrare un nuovo server remoto
type CreateSSHServerRequest struct {
	HostKeyFingerprint string             `json:"host_key_fingerprint"`
	ClientName         string             `json:"client_name"`
	Name               string             `json:"name"`
	Host               string             `json:"host"`
	Port               int                `json:"port"`
	Username           string             `json:"username"`
	AuthType           domain.SSHAuthType `json:"auth_type"`
	RawSecret          string             `json:"raw_secret"`
}

// UpdateSSHServerRequest racchiude i parametri per aggiornare un server remoto esistente
type UpdateSSHServerRequest struct {
	HostKeyFingerprint *string            `json:"host_key_fingerprint,omitempty"`
	ClientName         *string            `json:"client_name,omitempty"`
	Name               string             `json:"name"`
	Host               string             `json:"host"`
	Port               int                `json:"port"`
	Username           string             `json:"username"`
	AuthType           domain.SSHAuthType `json:"auth_type"`
	RawSecret          *string            `json:"raw_secret,omitempty"`
}

// SSHExecutor definisce l'interfaccia per l'esecuzione di comandi remoti SSH
type SSHExecutor interface {
	ExecuteCommand(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (stdout string, stderr string, exitCode int, err error)
}

// SSHService astrae la gestione dell'inventario dei server SSH e l'interazione sicura via RFC 4251
type SSHService interface {
	SSHExecutor
	CreateServer(ctx context.Context, tenantID uuid.UUID, req CreateSSHServerRequest) (*domain.SSHServer, error)
	GetServer(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.SSHServer, error)
	ListServers(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error)
	UpdateServer(ctx context.Context, tenantID, serverID uuid.UUID, req UpdateSSHServerRequest) (*domain.SSHServer, error)
	DeleteServer(ctx context.Context, tenantID, serverID uuid.UUID) error
	TestConnection(ctx context.Context, tenantID, serverID uuid.UUID) (*SSHConnectionTestResult, error)
	GetRawSSHClient(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error)
}

type sshService struct {
	repo      domain.SSHServerRepository
	masterKey []byte
}

// NewSSHService crea un'istanza di SSHService iniettando il repository e la master key AES-256
func NewSSHService(repo domain.SSHServerRepository, masterKey string) SSHService {
	return &sshService{
		repo:      repo,
		masterKey: []byte(masterKey),
	}
}

// CreateServer cifra il secret fornito tramite AES-256-GCM e persiste il nuovo server
func (s *sshService) CreateServer(ctx context.Context, tenantID uuid.UUID, req CreateSSHServerRequest) (*domain.SSHServer, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}
	if strings.TrimSpace(req.RawSecret) == "" {
		return nil, domain.ErrEmptySecret
	}
	if req.Port <= 0 {
		req.Port = 22
	}

	tenantKey, err := crypto.DeriveTenantKey(s.masterKey, tenantID)
	if err != nil {
		return nil, fmt.Errorf("derivazione tenant key fallita: %w", err)
	}

	encryptedSecret, err := crypto.EncryptAES256GCM([]byte(req.RawSecret), tenantKey, []byte(tenantID.String()))
	if err != nil {
		return nil, fmt.Errorf("cifratura credenziale SSH fallita: %w", err)
	}

	var fingerprint *string
	if req.AuthType == domain.SSHAuthTypeKey {
		if fp, err := extractPublicKeyFingerprint([]byte(req.RawSecret)); err == nil {
			fingerprint = &fp
		}
	}

	server := &domain.SSHServer{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		Name:                 strings.TrimSpace(req.Name),
		ClientName:           strings.TrimSpace(req.ClientName),
		Host:                 strings.TrimSpace(req.Host),
		Port:                 req.Port,
		Username:             strings.TrimSpace(req.Username),
		AuthType:             req.AuthType,
		EncryptedSecret:      encryptedSecret,
		PublicKeyFingerprint: fingerprint,
		HostKeyFingerprint:   strings.TrimSpace(req.HostKeyFingerprint),
		Status:               domain.SSHServerStatusUnchecked,
	}

	if err := server.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, server); err != nil {
		return nil, err
	}

	return server, nil
}

// GetServer recupera il server verificando l'appartenenza al tenant
func (s *sshService) GetServer(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.SSHServer, error) {
	return s.repo.GetByID(ctx, tenantID, serverID)
}

// ListServers elenca i server associati al tenant
func (s *sshService) ListServers(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error) {
	return s.repo.List(ctx, tenantID)
}

// UpdateServer aggiorna i metadati del server o ne rinnova il secret crittografato
func (s *sshService) UpdateServer(ctx context.Context, tenantID, serverID uuid.UUID, req UpdateSSHServerRequest) (*domain.SSHServer, error) {
	server, err := s.repo.GetByID(ctx, tenantID, serverID)
	if err != nil {
		return nil, err
	}

	server.Name = strings.TrimSpace(req.Name)
	endpointChanged := server.Host != strings.TrimSpace(req.Host) || (req.Port > 0 && req.Port != server.Port)
	if req.HostKeyFingerprint != nil {
		pin := strings.TrimSpace(*req.HostKeyFingerprint)
		if pin != server.HostKeyFingerprint {
			server.Status = domain.SSHServerStatusUnchecked
			server.LastConnectedAt = nil
		}
		server.HostKeyFingerprint = pin
	} else if endpointChanged {
		server.HostKeyFingerprint = ""
	}
	if endpointChanged {
		server.Status = domain.SSHServerStatusUnchecked
		server.LastConnectedAt = nil
	}
	if req.ClientName != nil {
		server.ClientName = strings.TrimSpace(*req.ClientName)
	}
	server.Host = strings.TrimSpace(req.Host)
	if req.Port > 0 {
		server.Port = req.Port
	}
	server.Username = strings.TrimSpace(req.Username)
	server.AuthType = req.AuthType

	if req.RawSecret != nil && strings.TrimSpace(*req.RawSecret) != "" {
		tenantKey, err := crypto.DeriveTenantKey(s.masterKey, tenantID)
		if err != nil {
			return nil, fmt.Errorf("derivazione tenant key fallita: %w", err)
		}
		encrypted, err := crypto.EncryptAES256GCM([]byte(*req.RawSecret), tenantKey, []byte(tenantID.String()))
		if err != nil {
			return nil, fmt.Errorf("cifratura nuova credenziale fallita: %w", err)
		}
		server.EncryptedSecret = encrypted

		if req.AuthType == domain.SSHAuthTypeKey {
			if fp, err := extractPublicKeyFingerprint([]byte(*req.RawSecret)); err == nil {
				server.PublicKeyFingerprint = &fp
			}
		} else {
			server.PublicKeyFingerprint = nil
		}
	}

	if err := server.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, server); err != nil {
		return nil, err
	}

	return server, nil
}

// DeleteServer elimina il server SSH
func (s *sshService) DeleteServer(ctx context.Context, tenantID, serverID uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, serverID)
}

// TestConnection tenta la connessione SSH, misura la latenza e verifica i privilegi di esecuzione
func (s *sshService) TestConnection(ctx context.Context, tenantID, serverID uuid.UUID) (*SSHConnectionTestResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	server, err := s.repo.GetByID(ctx, tenantID, serverID)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	client, err := s.getSSHClientForServer(ctx, server)
	if err != nil {
		now := time.Now().UTC()
		_ = s.repo.UpdateStatus(ctx, tenantID, serverID, domain.SSHServerStatusOffline, &now)
		return &SSHConnectionTestResult{
			Online: false,
			Error:  err.Error(),
		}, nil
	}
	defer client.Close()
	stop := context.AfterFunc(ctx, func() { _ = client.Close() })
	defer stop()

	latency := time.Since(start).Milliseconds()

	session, err := client.NewSession()
	if err != nil {
		now := time.Now().UTC()
		_ = s.repo.UpdateStatus(ctx, tenantID, serverID, domain.SSHServerStatusOffline, &now)
		return &SSHConnectionTestResult{
			Online: false,
			Error:  fmt.Sprintf("apertura sessione fallita: %v", err),
		}, nil
	}
	defer session.Close()

	// Verifica sistema operativo e ID utente (0 = root)
	output, err := session.CombinedOutput("uname -srm 2>/dev/null; id -u 2>/dev/null")
	if err != nil {
		now := time.Now().UTC()
		_ = s.repo.UpdateStatus(ctx, tenantID, serverID, domain.SSHServerStatusOffline, &now)
		return &SSHConnectionTestResult{
			Online: false,
			Error:  fmt.Sprintf("esecuzione check comandi fallita: %v", err),
		}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	osInfo := "Linux"
	isRoot := false

	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		osInfo = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 && strings.TrimSpace(lines[1]) == "0" {
		isRoot = true
	}

	now := time.Now().UTC()
	_ = s.repo.UpdateStatus(ctx, tenantID, serverID, domain.SSHServerStatusOnline, &now)

	return &SSHConnectionTestResult{
		Online:    true,
		LatencyMs: latency,
		OSInfo:    osInfo,
		IsRoot:    isRoot,
	}, nil
}

// ExecuteCommand esegue un comando remoto via SSH catturando stdout, stderr ed exit code
func (s *sshService) ExecuteCommand(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	server, err := s.repo.GetByID(ctx, tenantID, serverID)
	if err != nil {
		return "", "", -1, err
	}

	client, err := s.getSSHClientForServer(ctx, server)
	if err != nil {
		return "", "", -1, fmt.Errorf("connessione SSH fallita: %w", err)
	}
	defer client.Close()
	stop := context.AfterFunc(ctx, func() { _ = client.Close() })
	defer stop()

	session, err := client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("creazione sessione SSH fallita: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(cmd)
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	exitCode := 0
	if err != nil {
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitStatus()
		} else {
			exitCode = -1
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode, err
}

// GetRawSSHClient espone direttamente il client SSH nativo per WebSocket PTY streaming
func (s *sshService) GetRawSSHClient(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error) {
	server, err := s.repo.GetByID(ctx, tenantID, serverID)
	if err != nil {
		return nil, err
	}
	return s.getSSHClientForServer(ctx, server)
}

func (s *sshService) getSSHClientForServer(ctx context.Context, server *domain.SSHServer) (*ssh.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !domain.ValidHostKeyFingerprint(server.HostKeyFingerprint) {
		return nil, fmt.Errorf("%w: configurare il fingerprint SHA256 della chiave host tramite un canale fidato", domain.ErrValidation)
	}
	tenantKey, err := crypto.DeriveTenantKey(s.masterKey, server.TenantID)
	if err != nil {
		return nil, fmt.Errorf("derivazione tenant key fallita: %w", err)
	}

	rawSecret, err := crypto.DecryptAES256GCM(server.EncryptedSecret, tenantKey, []byte(server.TenantID.String()))
	if err != nil {
		return nil, fmt.Errorf("decifratura credenziale fallita: %w", err)
	}

	var authMethods []ssh.AuthMethod
	switch server.AuthType {
	case domain.SSHAuthTypePassword:
		authMethods = append(authMethods, ssh.Password(string(rawSecret)))
	case domain.SSHAuthTypeKey:
		signer, err := ssh.ParsePrivateKey(rawSecret)
		if err != nil {
			return nil, fmt.Errorf("parsing chiave privata SSH fallito: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	default:
		return nil, domain.ErrInvalidSSHAuthType
	}

	clientConfig := &ssh.ClientConfig{
		User:    server.Username,
		Auth:    authMethods,
		Timeout: 10 * time.Second,
	}

	addr := net.JoinHostPort(server.Host, strconv.Itoa(server.Port))
	handshakeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Match the defaults in x/crypto/ssh v0.27.0. Remove a rejected key family
	// (including all RSA signature variants) before opening another connection.
	algorithms := []string{
		ssh.CertAlgoRSASHA256v01, ssh.CertAlgoRSASHA512v01,
		ssh.CertAlgoRSAv01, ssh.CertAlgoDSAv01, ssh.CertAlgoECDSA256v01,
		ssh.CertAlgoECDSA384v01, ssh.CertAlgoECDSA521v01, ssh.CertAlgoED25519v01,
		ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521,
		ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSASHA512,
		ssh.KeyAlgoRSA, ssh.KeyAlgoDSA, ssh.KeyAlgoED25519,
	}
	var observed []string
	for len(algorithms) > 0 {
		var rejectedType string
		clientConfig.HostKeyAlgorithms = algorithms
		clientConfig.HostKeyCallback = func(_ string, _ net.Addr, key ssh.PublicKey) error {
			fingerprint := ssh.FingerprintSHA256(key)
			if fingerprint == server.HostKeyFingerprint {
				return nil
			}
			rejectedType = key.Type()
			observed = append(observed, key.Type()+" "+fingerprint)
			return errors.New("chiave host SSH diversa dal fingerprint autorizzato")
		}
		client, err := dialPinnedSSHClient(handshakeCtx, addr, clientConfig)
		if err == nil {
			return client, nil
		}
		if handshakeCtx.Err() != nil {
			return nil, handshakeCtx.Err()
		}
		if rejectedType == "" {
			if len(observed) == 0 {
				return nil, err
			}
			return nil, fmt.Errorf("%w; chiave host SSH diversa dal fingerprint autorizzato: atteso %s; osservate %s. Verificare il pin tramite un canale fidato", err, server.HostKeyFingerprint, strings.Join(observed, ", "))
		}
		remaining := make([]string, 0, len(algorithms))
		for _, algorithm := range algorithms {
			keyType := algorithm
			switch algorithm {
			case ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSASHA512:
				keyType = ssh.KeyAlgoRSA
			case ssh.CertAlgoRSASHA256v01, ssh.CertAlgoRSASHA512v01:
				keyType = ssh.CertAlgoRSAv01
			}
			if keyType != rejectedType {
				remaining = append(remaining, algorithm)
			}
		}
		algorithms = remaining
	}
	return nil, fmt.Errorf("chiave host SSH diversa dal fingerprint autorizzato: atteso %s; osservate %s. Verificare il pin tramite un canale fidato", server.HostKeyFingerprint, strings.Join(observed, ", "))
}

// All attempts share the caller's handshake deadline; credentials are sent only
// after HostKeyCallback accepts the exact authorized fingerprint.
func dialPinnedSSHClient(ctx context.Context, addr string, clientConfig *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s fallito: %w", addr, err)
	}
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()
		return nil, err
	}
	closed := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { _ = conn.Close(); close(closed) })
	c, channels, requests, err := ssh.NewClientConn(conn, addr, clientConfig)
	if !stop() {
		<-closed
	}
	if ctx.Err() != nil {
		_ = conn.Close()
		return nil, ctx.Err()
	}
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("handshake SSH fallito: %w", err)
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		_ = c.Close()
		return nil, err
	}
	return ssh.NewClient(c, channels, requests), nil
}

func extractPublicKeyFingerprint(privateKeyPEM []byte) (string, error) {
	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(signer.PublicKey()), nil
}
