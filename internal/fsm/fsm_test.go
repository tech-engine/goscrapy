package fsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	key,
	val,
	op string
}

func TestFsmGet(t *testing.T) {
	fsm := New[string, string](5)
	fsm.Set("KEY_1", "VAL_1")

	val, ok := fsm.Get("KEY_1")

	assert.True(t, ok, "ok=true expected for key=%s", "KEY_1")

	assert.Equalf(t, "VAL_1", val, "val=%s expected for key=%s but got %s", "VAL_1", "KEY_1", val)

	val, ok = fsm.Get("KEY_2")

	assert.Falsef(t, ok, "ok=false expected for key=%s", "KEY_2")

	assert.Emptyf(t, val, "empty string expected for key=%s but got %s")
}

func TestFsmSet(t *testing.T) {
	fsm := New[string, string](5)
	err := fsm.Set("KEY_1", "VAL_1")

	assert.NoErrorf(t, err, "nil error expected but got %s", err)
}

func TestFsmLimit(t *testing.T) {
	var err error
	fsm := New[string, string](2)

	err = fsm.Set("KEY_1", "VAL_1")
	assert.NoErrorf(t, err, "nil error expected for key=%s but got %s", "KEY_1", err)

	err = fsm.Set("KEY_2", "VAL_2")
	assert.NoErrorf(t, err, "nil error expected for key=%s but got %s", "KEY_2", err)

	err = fsm.Set("KEY_3", "VAL_3")

	if assert.Errorf(t, err, "error is expected for key=%s", "KEY_3") {
		assert.ErrorIs(t, err, ERR_MAX_ITEM_EXCEEDED)
	}
}

func TestFsmClear(t *testing.T) {
	fsm := New[string, string](5)

	fsm.Set("KEY_1", "VAL_1")

	fsm.Clear()

	val, ok := fsm.Get("KEY_1")

	assert.Falsef(t, ok, "ok=false expected for key=%s", "KEY_1")

	assert.Emptyf(t, val, "empty string expected for key=%s but got %s")
}
