package migration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/migration"
)

func TestApplyError_Error_success(t *testing.T) {
	applyErr := &migration.ApplyError{}

	want := migration.ErrFailedToApply.Error()
	got := applyErr.Error()

	assert.Equal(t, want, got)
}

func TestApplyError_Error_successWhenFileAndCodeAreSet(t *testing.T) {
	applyErr := &migration.ApplyError{File: "001_a.sql", Code: "42601"}

	want := "migration: failed to apply migrations: file 001_a.sql: postgresql error 42601"
	got := applyErr.Error()

	assert.Equal(t, want, got)
}

func TestApplyError_Error_successWhenOnlyFileIsSet(t *testing.T) {
	applyErr := &migration.ApplyError{File: "001_a.sql"}

	want := "migration: failed to apply migrations: file 001_a.sql"
	got := applyErr.Error()

	assert.Equal(t, want, got)
}

func TestApplyError_Error_successWhenAttributionIsUnavailableWithCode(t *testing.T) {
	applyErr := &migration.ApplyError{AttributionUnavailable: true, Code: "42601"}

	want := "migration: failed to apply migrations: file attribution unavailable: postgresql error 42601"
	got := applyErr.Error()

	assert.Equal(t, want, got)
}

func TestApplyError_Error_successWhenAttributionIsUnavailable(t *testing.T) {
	applyErr := &migration.ApplyError{AttributionUnavailable: true}

	want := "migration: failed to apply migrations: file attribution unavailable"
	got := applyErr.Error()

	assert.Equal(t, want, got)
}

func TestApplyError_Error_successWhenOnlyCodeIsSet(t *testing.T) {
	applyErr := &migration.ApplyError{Code: "42601"}

	want := "migration: failed to apply migrations: postgresql error 42601"
	got := applyErr.Error()

	assert.Equal(t, want, got)
}

func TestApplyError_Unwrap_success(t *testing.T) {
	cause := errors.New("execution error")
	mockRunner := migration.MockRunner(cause)
	mockStatus := migration.MockStatus(nil, errors.New("status failed"))
	apply := migration.MakeApply(mockRunner, mockStatus)
	var applyErr *migration.ApplyError
	_ = errors.As(apply(context.Background()), &applyErr)

	want := []error{migration.ErrFailedToApply, cause}
	got := applyErr.Unwrap()

	assert.Equal(t, want, got)
}
