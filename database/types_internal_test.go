package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveDatabaseURL_success(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "defaults to the compose service name when the host is unset",
			host: "",
			want: "postgresql://user:pass@postgres_db:5432/app?sslmode=disable",
		},
		{
			name: "uses POSTGRES_DB_HOST when set",
			host: "nebula-postgres",
			want: "postgresql://user:pass@nebula-postgres:5432/app?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTGRES_DB_USER", "user")
			t.Setenv("POSTGRES_DB_PASS", "pass")
			t.Setenv("POSTGRES_DB_NAME", "app")
			t.Setenv("POSTGRES_DB_PORT", "5432")
			t.Setenv("POSTGRES_DB_HOST", tt.host)

			assert.Equal(t, tt.want, resolveDatabaseURL())
		})
	}
}

func TestResolveAdminURL_success(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "secret")
	t.Setenv("POSTGRES_DB_HOST", "")
	t.Setenv("POSTGRES_DB_PORT", "5432")

	want := "postgresql://admin:secret@postgres_db:5432/postgres?sslmode=disable"
	got, _ := resolveAdminURL()

	assert.Equal(t, want, got)
}

func TestResolveAdminURL_failsWhenCredentialsAreMissing(t *testing.T) {
	t.Setenv("POSTGRES_ADMIN_USER", "admin")
	t.Setenv("POSTGRES_ADMIN_PASS", "")

	_, got := resolveAdminURL()

	assert.False(t, got)
}

func TestResolveDatabaseURL_successWhenPasswordHasSpecialCharacters(t *testing.T) {
	t.Setenv("POSTGRES_DB_USER", "user")
	t.Setenv("POSTGRES_DB_PASS", "p w=1@/:")
	t.Setenv("POSTGRES_DB_HOST", "host")
	t.Setenv("POSTGRES_DB_NAME", "app")
	t.Setenv("POSTGRES_DB_PORT", "5432")

	want := "postgresql://user:p%20w=1%40%2F%3A@host:5432/app?sslmode=disable"
	got := resolveDatabaseURL()

	assert.Equal(t, want, got)
}
