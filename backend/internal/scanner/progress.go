package scanner

import (
	"bytes"
	"encoding/xml"
	"math"
)

// TaskProgress describes a single Nmap task, not the entire scan.
type TaskProgress struct {
	Task      string   `xml:"task,attr"`
	Percent   *float64 `xml:"percent,attr"`
	Remaining *float64 `xml:"remaining,attr"`
}

// progressWriter observes bounded XML fragments; the original output is retained separately.
type progressWriter struct {
	pending    []byte
	onProgress func(TaskProgress)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	length := len(p)
	// Process bounded slices so even a very large stdout write cannot grow this buffer unboundedly.
	for len(p) > 0 {
		n := len(p)
		if n > 4096 {
			n = 4096
		}
		w.pending = append(w.pending, p[:n]...)
		p = p[n:]
		for {
			start := bytes.Index(w.pending, []byte("<taskprogress "))
			if start < 0 {
				if len(w.pending) > 32 {
					w.pending = w.pending[len(w.pending)-32:]
				}
				break
			}
			w.pending = w.pending[start:]
			end := bytes.IndexByte(w.pending, '>')
			if end < 0 {
				if len(w.pending) > 8192 {
					w.pending = nil
				}
				break
			}
			var progress TaskProgress
			if xml.Unmarshal(w.pending[:end+1], &progress) == nil && progress.Percent != nil && !math.IsNaN(*progress.Percent) && !math.IsInf(*progress.Percent, 0) && *progress.Percent >= 0 && *progress.Percent <= 100 {
				if progress.Remaining != nil && (math.IsNaN(*progress.Remaining) || math.IsInf(*progress.Remaining, 0) || *progress.Remaining < 0) {
					progress.Remaining = nil
				}
				w.onProgress(progress)
			}
			w.pending = w.pending[end+1:]
		}
	}
	return length, nil
}
