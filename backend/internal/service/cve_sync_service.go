package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service/fixtures"
)

// SyncStats racchiude le metriche dell'operazione di sincronizzazione
type SyncStats struct {
	TotalProcessed int           `json:"total_processed"`
	TotalUpserted  int           `json:"total_upserted"`
	BatchCount     int           `json:"batch_count"`
	Duration       time.Duration `json:"duration"`
}

// CVESyncService gestisce l'importazione a streaming di feed CVE con inserimento a batch
type CVESyncService struct {
	cveRepo   domain.CVERepository
	batchSize int
}

// NewCVESyncService inizializza una nuova istanza di CVESyncService
func NewCVESyncService(cveRepo domain.CVERepository, batchSize int) *CVESyncService {
	if batchSize <= 0 {
		batchSize = 500
	}
	return &CVESyncService{
		cveRepo:   cveRepo,
		batchSize: batchSize,
	}
}

// SyncFromReader legge un flusso JSON di vulnerabilità CVE da un io.Reader arbitrario
// ed esegue l'upsert a blocchi nel database PostgreSQL senza picchi di memoria (NFR-PERF-02, NFR-PERF-03).
func (s *CVESyncService) SyncFromReader(ctx context.Context, reader io.Reader) (*SyncStats, error) {
	if reader == nil {
		return nil, errors.New("reader non valido (nil)")
	}
	if s.cveRepo == nil {
		return nil, errors.New("cve repository non configurato")
	}

	startTime := time.Now()
	stats := &SyncStats{}

	dec := json.NewDecoder(reader)

	// Leggi il primo token per verificare se si tratta di un array JSON '['
	tok, err := dec.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return stats, nil
		}
		return nil, fmt.Errorf("lettura token iniziale JSON fallita: %w", err)
	}

	isArray := false
	if delim, ok := tok.(json.Delim); ok && delim == '[' {
		isArray = true
	}

	batch := make([]*domain.CVEEntry, 0, s.batchSize)
	now := time.Now().UTC()

	processEntry := func(entry *domain.CVEEntry) error {
		if entry.CVEID == "" {
			return nil // Salta voci prive di identificativo
		}
		stats.TotalProcessed++

		// Arricchimento campi mancanti
		if entry.Severity == "" {
			if entry.CVSSScore != nil {
				entry.Severity = domain.SeverityFromCVSS(*entry.CVSSScore)
			} else {
				entry.Severity = domain.SeverityInfo
			}
		}
		if entry.UpdatedAt.IsZero() {
			entry.UpdatedAt = now
		}
		if entry.CPEMatchCriteria == nil {
			entry.CPEMatchCriteria = []domain.CPEMatchCriterion{}
		}

		batch = append(batch, entry)

		// Raggiunta la soglia di batch: esecuzione upsert
		if len(batch) >= s.batchSize {
			if err := s.flushBatch(ctx, batch); err != nil {
				return err
			}
			stats.TotalUpserted += len(batch)
			stats.BatchCount++
			batch = batch[:0]
		}

		return nil
	}

	if isArray {
		for dec.More() {
			if ctx.Err() != nil {
				return stats, ctx.Err()
			}
			var entry domain.CVEEntry
			if err := dec.Decode(&entry); err != nil {
				return stats, fmt.Errorf("decodifica elemento CVE #%d fallita: %w", stats.TotalProcessed+1, err)
			}
			if err := processEntry(&entry); err != nil {
				return stats, err
			}
		}

		// Consuma il token di chiusura ']'
		if _, err := dec.Token(); err != nil && !errors.Is(err, io.EOF) {
			return stats, fmt.Errorf("lettura token di chiusura array fallita: %w", err)
		}
	} else {
		// Decodifica primo elemento
		var firstEntry domain.CVEEntry
		// Nel caso il token iniziale non fosse un delimitatore, decodifichiamo lo stream riga per riga / oggetto per oggetto
		for {
			if ctx.Err() != nil {
				return stats, ctx.Err()
			}
			var entry domain.CVEEntry
			if err := dec.Decode(&entry); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return stats, fmt.Errorf("decodifica voce CVE streaming fallita: %w", err)
			}
			if err := processEntry(&entry); err != nil {
				return stats, err
			}
		}
		_ = firstEntry
	}

	// Flush finale degli elementi residui
	if len(batch) > 0 {
		if err := s.flushBatch(ctx, batch); err != nil {
			return stats, err
		}
		stats.TotalUpserted += len(batch)
		stats.BatchCount++
		batch = batch[:0]
	}

	stats.Duration = time.Since(startTime)
	return stats, nil
}

// SyncDefaultFeed importa il dataset CVE di default incorporato a tempo di compilazione
func (s *CVESyncService) SyncDefaultFeed(ctx context.Context) (*SyncStats, error) {
	file, err := fixtures.DefaultFeedFS.Open("cve_sample_feed.json")
	if err != nil {
		return nil, fmt.Errorf("apertura feed CVE predefinito incorporato fallita: %w", err)
	}
	defer file.Close()

	return s.SyncFromReader(ctx, file)
}

func (s *CVESyncService) flushBatch(ctx context.Context, batch []*domain.CVEEntry) error {
	if len(batch) == 0 {
		return nil
	}
	if err := s.cveRepo.UpsertBatch(ctx, batch); err != nil {
		return fmt.Errorf("upsert batch di %d CVE fallito: %w", len(batch), err)
	}
	return nil
}
