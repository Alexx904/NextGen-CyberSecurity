package report

import (
	"bytes"
	"encoding/csv"
	"sort"
	"strings"

	"nextgen-cybersec/internal/domain"
)

// GenerateDepartmentCSV keeps the scan CSV contract and appends department for consolidated exports.
func GenerateDepartmentCSV(findings []*domain.Finding, departments map[string]string) ([]byte, error) {
	data, err := GenerateCSV(findings)
	if err != nil {
		return nil, err
	}
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return nil, err
	}
	rows[0] = append(rows[0], "reparto")
	for i := 1; i < len(rows); i++ {
		name := departments[rows[i][1]]
		if name == "" {
			name = "Non attribuito"
		}
		if strings.ContainsAny(name[:1], "=+-@\t\r") {
			name = "'" + name
		}
		rows[i] = append(rows[i], name)
	}
	sort.SliceStable(rows[1:], func(i, j int) bool {
		a, b := rows[i+1], rows[j+1]
		if a[len(a)-1] != b[len(b)-1] {
			return a[len(a)-1] < b[len(b)-1]
		}
		return a[2] < b[2]
	})
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	if err = writer.WriteAll(rows); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
