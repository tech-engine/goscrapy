package cmap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	key,
	val,
	op string
}

func TestCMapGet(t *testing.T) {
	cmap := NewCMap[string, string]()
	cmap.Set("KEY_1", "VAL_1")

	val, ok := cmap.Get("KEY_1")

	assert.True(t, ok, "ok=true expected for key=%s", "KEY_1")

	assert.Equalf(t, "VAL_1", val, "val=%s expected for key=%s but got %s", "VAL_1", "KEY_1", val)

	val, ok = cmap.Get("KEY_2")

	assert.Falsef(t, ok, "ok=false expected for key=%s", "KEY_2")

	assert.Emptyf(t, val, "empty string expected for key=%s but got %s")
}

func TestCMapSet(t *testing.T) {
	cmap := NewCMap[string, string]()
	err := cmap.Set("KEY_1", "VAL_1")

	assert.NoErrorf(t, err, "nil error expected but got %s", err)
}

func TestCMapLimit(t *testing.T) {
	var err error
	cmap := NewCMap[string, string](WithSize(2))

	err = cmap.Set("KEY_1", "VAL_1")
	assert.NoErrorf(t, err, "nil error expected for key=%s but got %s", "KEY_1", err)

	err = cmap.Set("KEY_2", "VAL_2")
	assert.NoErrorf(t, err, "nil error expected for key=%s but got %s", "KEY_2", err)

	err = cmap.Set("KEY_3", "VAL_3")

	if assert.Errorf(t, err, "error is expected for key=%s", "KEY_3") {
		assert.ErrorIs(t, err, ERR_MAX_ITEM_EXCEEDED)
	}

	cmap.Del("KEY_2")

	err = cmap.Set("KEY_3", "VAL_3")

	assert.NoErrorf(t, err, "nil error expected for key=%s but got %s", "KEY_3", err)
}

func TestCMapDel(t *testing.T) {
	cmap := NewCMap[string, string]()

	cmap.Set("KEY_1", "VAL_1")

	cmap.Del("KEY_1")

	assert.Equalf(t, 0, cmap.Len(), "len=%d expected but has %d", 0, cmap.Len())

	val, ok := cmap.Get("KEY_1")

	assert.Falsef(t, ok, "false expected for key=%s", "KEY_1")
	assert.Equalf(t, "", val, "empty string expected but got %s", val)
}

func TestCMapConcurrency(t *testing.T) {

	testCases := []testCase{
		{
			key: "KEY_1",
			val: "VAL_1",
			op:  "WRITE",
		},
		{
			key: "KEY_2",
			val: "VAL_2",
			op:  "WRITE",
		},
		{
			key: "KEY_2",
			val: "VAL_2",
			op:  "READ",
		},
		{
			key: "KEY_3",
			val: "VAL_3",
			op:  "WRITE",
		},
		{
			key: "KEY_4",
			val: "VAL_4",
			op:  "WRITE",
		},
		{
			key: "KEY_1",
			val: "VAL_1",
			op:  "READ",
		},
		{
			key: "KEY_1",
			val: "VAL_1",
			op:  "WRITE",
		},
	}

	cmap := NewCMap[string, string](WithSize(5))

	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			for _, tc := range testCases {
				if tc.op == "READ" {
					cmap.Get(tc.key)
				} else {
					err := cmap.Set(tc.key, tc.val)

					assert.NoError(t, err)
				}
			}
		})
	}
}
