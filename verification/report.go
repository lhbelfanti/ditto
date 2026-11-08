package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteReport saves report as report.json in dir, and creates dir when it does not exist.
func WriteReport(dir string, report Report) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCantCreateDir, err)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCantMarshalReport, err)
	}

	err = os.WriteFile(filepath.Join(dir, "report.json"), data, 0o644)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCantWriteReport, err)
	}

	return nil
}
