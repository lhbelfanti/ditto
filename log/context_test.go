package log_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
)

func TestParam_success(t *testing.T) {
	want := "key"
	got := log.Param("key", "value").Key

	assert.Equal(t, want, got)
}

func TestParam_successWhenValueIsKept(t *testing.T) {
	want := "value"
	got := log.Param("key", "value").Value

	assert.Equal(t, want, got)
}

func TestWith_successWhenChildAddsFieldsParentDoesNotSeeThem(t *testing.T) {
	buf := log.MockLogOutput(t)
	parent := log.With(context.Background(), log.Param("request_id", "abc"))
	_ = log.With(parent, log.Param("worker", 1))
	log.Info(parent, "from the parent")

	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)

	got := entry["worker"]

	assert.Nil(t, got)
}

func TestWith_successWhenDerivedConcurrentlyFromTheSameParent(t *testing.T) {
	parent := log.With(context.Background(), log.Param("request_id", "abc"))
	var wg sync.WaitGroup

	for i := range 8 {
		wg.Go(func() {
			_ = log.With(parent, log.Param("worker", i))
		})
	}
	wg.Wait()

	buf := log.MockLogOutput(t)
	log.Info(parent, "from the parent")

	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)

	got := entry["worker"]

	assert.Nil(t, got)
}
