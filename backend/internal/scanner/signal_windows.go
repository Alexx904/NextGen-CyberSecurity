//go:build windows

package scanner

import (
	"os"
)

// defaultCancelSignal definisce il segnale inviato al processo Nmap su ambienti Windows
var defaultCancelSignal os.Signal = os.Interrupt

