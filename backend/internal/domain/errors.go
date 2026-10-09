package domain

import "errors"

var (
	// Errori generici di dominio
	ErrNotFound      = errors.New("risorsa non trovata")
	ErrConflict      = errors.New("conflitto con risorsa esistente")
	ErrValidation    = errors.New("errore di validazione dei dati")
	ErrUnauthorized  = errors.New("autenticazione richiesta o credenziali non valide")
	ErrForbidden     = errors.New("permessi insufficienti per accedere alla risorsa")

	// Errori specifici per Target
	ErrTargetBlacklisted = errors.New("il target è presente nella blacklist e non può essere scansionato")
	ErrTargetInvalid     = errors.New("formato del target non valido (IP, CIDR, Range o FQDN atteso)")

	// Errori specifici per Scansioni e Schedulazioni
	ErrScanNotFound           = errors.New("scansione non trovata")
	ErrScanAlreadyRunning     = errors.New("una scansione su questo target è già in corso")
	ErrScanCannotBeCancelled  = errors.New("la scansione non si trova in uno stato annullabile")
	ErrInvalidStateTransition = errors.New("transizione di stato della scansione non consentita")
	ErrScheduleInactive       = errors.New("la pianificazione richiesta è disattivata")
	ErrInvalidCron            = errors.New("espressione cron non valida")

	// Errori specifici per Nmap Scanner
	ErrDisallowedScanOption = errors.New("argomento di scansione non consentito o non sicuro")
	ErrScannerExecFailed    = errors.New("esecuzione del comando di scansione fallita")
	ErrScanTimeout          = errors.New("tempo massimo per l'esecuzione della scansione superato")

	// Errori specifici per Server SSH e Firewall
	ErrSSHServerNotFound      = errors.New("server SSH non trovato")
	ErrSSHServerDuplicateName = errors.New("un server SSH con questo nome esiste già per il tenant")
	ErrInvalidSSHAuthType     = errors.New("tipo di autenticazione SSH non supportato (atteso KEY o PASSWORD)")
	ErrInvalidSSHPort         = errors.New("porta SSH non valida (range consentito 1-65535)")
	ErrFirewallBackupNotFound = errors.New("backup firewall non trovato")
	ErrEmptyHost              = errors.New("l'indirizzo host del server SSH non può essere vuoto")
	ErrEmptyUsername          = errors.New("lo username SSH non può essere vuoto")
	ErrEmptySecret            = errors.New("il secret o credenziale crittografata SSH non può essere vuoto")
	ErrInvalidChainName       = errors.New("nome catena firewall non valido (atteso INPUT, OUTPUT o FORWARD)")
	ErrInvalidRuleAction      = errors.New("azione regola firewall non valida (atteso ACCEPT, DROP, REJECT o LOG)")
	ErrSafetyTimerActive      = errors.New("un safety timer di configurazione firewall è già attivo per questo server")
	ErrNoSafetyTimerActive    = errors.New("nessun safety timer di configurazione firewall attivo per questo server")
	ErrSSHExecutionFailed     = errors.New("esecuzione del comando remoto SSH fallita")
)

