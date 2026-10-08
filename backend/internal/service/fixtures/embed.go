package fixtures

import "embed"

// DefaultFeedFS incorpora il dataset JSON delle CVE di base per l'avvio rapido e gli ambienti offline (Air-Gapped)
//
//go:embed *.json
var DefaultFeedFS embed.FS
