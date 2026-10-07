package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen_success(t *testing.T) {
	pg, _ := open(context.Background(), "postgresql://u:p@127.0.0.1:5432/d?sslmode=disable")
	defer pg.Close()

	got := pg.Database()

	assert.NotNil(t, got)
}

func TestOpen_failsWhenURLIsMalformed(t *testing.T) {
	want := ErrCantInitDatabase
	_, got := open(context.Background(), "postgresql://u:p@127.0.0.1:not-a-port/d")

	assert.ErrorIs(t, got, want)
}
