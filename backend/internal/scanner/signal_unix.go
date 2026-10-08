//go:build !windows

package scanner

import (
	"os"
	"syscall"
)

// defaultCancelSignal definisce il segnale inviato al processo Nmap all'annullamento del context (NFR-REL-01)
var defaultCancelSignal os.Signal = syscall.SIGTERM

