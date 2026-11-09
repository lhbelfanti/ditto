package verification

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Table returns the table of a feature card for steps: one row per step, with its request, its status and what
// it proves.
func Table(steps []Step) string {
	var table strings.Builder
	table.WriteString("| # | Request | Status | Proves |\n|---|---|---|---|\n")
	for i, step := range steps {
		fmt.Fprintf(&table, "| %d | `%s` | %d | %s |\n", i+1, step.Request(), step.Status, step.Description)
	}

	return table.String()
}

// UpdateCards writes the table of each journey in the feature cards of dir, under the "### <ID>" heading of the
// journey, replacing the table that is there. Every journey must have a heading with a table in one of the cards.
func UpdateCards(dir string, entries []Entry) error {
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCantListCards, err)
	}

	tables := make(map[string]string, len(entries))
	for _, entry := range entries {
		tables[entry.ID] = Table(entry.Journey.Steps())
	}

	updated := make(map[string]bool, len(entries))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCantReadCard, err)
		}

		card, ids := replaceTables(string(data), tables)
		for _, id := range ids {
			updated[id] = true
		}

		if card == string(data) {
			continue
		}

		err = os.WriteFile(path, []byte(card), 0o644)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCantWriteCard, err)
		}
	}

	for _, entry := range entries {
		if !updated[entry.ID] {
			return fmt.Errorf("%w: %s", ErrCardNotFound, entry.ID)
		}
	}

	return nil
}

// replaceTables replaces, in card, the first table under each "### <ID>" heading that has a table in tables, and
// returns the card and the IDs whose table it replaced. A heading followed by another heading before any table
// keeps no table.
func replaceTables(card string, tables map[string]string) (string, []string) {
	lines := strings.Split(card, "\n")
	written := make([]string, 0, len(lines))
	var ids []string
	id := ""
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#") {
			id = headingID(line)
		}

		table, wanted := tables[id]
		if !wanted || !strings.HasPrefix(line, "|") {
			written = append(written, line)
			continue
		}

		for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "|") {
			i++
		}

		written = append(written, strings.Split(strings.TrimSuffix(table, "\n"), "\n")...)
		ids = append(ids, id)
		id = ""
	}

	return strings.Join(written, "\n"), ids
}

// headingID returns the journey ID of a "### <ID> — <title>" heading, or "" for any other line.
func headingID(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != "###" {
		return ""
	}

	return fields[1]
}
