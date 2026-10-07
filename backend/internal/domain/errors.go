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
)

