package verification_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestTable_success(t *testing.T) {
	mockSteps := []verification.Step{
		{Description: "lists the sessions", Method: http.MethodGet, Path: "/accounts/me/sessions/v1", Status: http.StatusOK},
		{Description: "revokes the other session", Method: http.MethodDelete, Path: "/accounts/me/sessions/{id}/v1", Status: http.StatusOK},
	}

	want := "| # | Request | Status | Proves |\n|---|---|---|---|\n" +
		"| 1 | `GET /accounts/me/sessions/v1` | 200 | lists the sessions |\n" +
		"| 2 | `DELETE /accounts/me/sessions/{id}/v1` | 200 | revokes the other session |\n"
	got := verification.Table(mockSteps)

	assert.Equal(t, want, got)
}

func TestUpdateCards_success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.md")
	_ = os.WriteFile(path, []byte(verification.MockCard("| # | Request |\n|---|---|\n| 1 | old |\n")), 0o644)
	mockStep := verification.Step{Description: "revokes the other session", Method: http.MethodDelete, Path: "/accounts/me/sessions/{id}/v1", Status: http.StatusOK}
	mockEntries := []verification.Entry{{ID: "session.revoke-one", Journey: verification.MockJourney(mockStep)}}

	err := verification.UpdateCards(dir, mockEntries)
	written, _ := os.ReadFile(path)

	want := []any{nil, verification.MockCard(verification.Table([]verification.Step{mockStep}))}
	got := []any{err, string(written)}

	assert.Equal(t, want, got)
}

func TestUpdateCards_failsWhenAJourneyHasNoCard(t *testing.T) {
	mockEntries := []verification.Entry{{ID: "session.revoke-one", Journey: verification.MockJourney()}}

	want := verification.ErrCardNotFound
	got := verification.UpdateCards(t.TempDir(), mockEntries)

	assert.ErrorIs(t, got, want)
}

func TestUpdateCards_failsWhenTheCardsCantBeListed(t *testing.T) {
	want := verification.ErrCantListCards
	got := verification.UpdateCards("[", nil)

	assert.ErrorIs(t, got, want)
}

func TestUpdateCards_failsWhenACardCantBeRead(t *testing.T) {
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "sessions.md"), 0o755)

	want := verification.ErrCantReadCard
	got := verification.UpdateCards(dir, nil)

	assert.ErrorIs(t, got, want)
}

func TestUpdateCards_failsWhenACardCantBeWritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.md")
	_ = os.WriteFile(path, []byte(verification.MockCard("| old |\n")), 0o444)
	mockEntries := []verification.Entry{{ID: "session.revoke-one", Journey: verification.MockJourney()}}

	want := verification.ErrCantWriteCard
	got := verification.UpdateCards(dir, mockEntries)

	assert.ErrorIs(t, got, want)
}
