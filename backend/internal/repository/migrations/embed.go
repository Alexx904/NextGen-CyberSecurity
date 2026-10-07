package migrations

import "embed"

// FS incorpora a tempo di compilazione tutti i file di migrazione SQL (*.sql)
// garantendo l'esecuzione zero-dependency all'avvio del container.
//
//go:embed *.sql
var FS embed.FS

