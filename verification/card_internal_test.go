package verification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceTables_success(t *testing.T) {
	card := "### a.one — One\n\nSetup: x.\n\n| old |\n| row |\n\n## Gotchas\n\n| kept |\n"
	tables := map[string]string{"a.one": "| new |\n|---|\n"}

	written, ids := replaceTables(card, tables)

	want := []any{"### a.one — One\n\nSetup: x.\n\n| new |\n|---|\n\n## Gotchas\n\n| kept |\n", []string{"a.one"}}
	got := []any{written, ids}

	assert.Equal(t, want, got)
}

func TestReplaceTables_successWhenTheHeadingHasNoTable(t *testing.T) {
	card := "### a.one — One\n\n### a.two — Two\n\n| kept |\n"
	tables := map[string]string{"a.one": "| new |\n"}

	written, ids := replaceTables(card, tables)

	want := []any{card, []string(nil)}
	got := []any{written, ids}

	assert.Equal(t, want, got)
}

func TestHeadingID_success(t *testing.T) {
	want := "session.revoke-one"
	got := headingID("### session.revoke-one — Revoke one session")

	assert.Equal(t, want, got)
}

func TestHeadingID_successWhenTheLineIsNotAJourneyHeading(t *testing.T) {
	want := ""
	got := headingID("## Gotchas")

	assert.Equal(t, want, got)
}
