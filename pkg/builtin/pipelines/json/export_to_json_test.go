package json

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport2JSON(t *testing.T) {
	f, err := os.CreateTemp(".", "export_test_*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	pipeline := New[*testutils.DummyRecord](Options{
		File:      f,
		Immediate: true,
	})
	defer pipeline.Close()

	require.NoError(t, pipeline.Open(context.Background()))

	record := &testutils.DummyRecord{Id: "1", Name: "rick"}
	assert.NoError(t, pipeline.ProcessItem(nil, record))

	f.Seek(0, 0)
	d := json.NewDecoder(f)
	var out testutils.DummyRecord
	err = d.Decode(&out)
	assert.NoError(t, err)
	assert.Equal(t, "1", out.Id)
}

func TestExport2JSON_Concurrent(t *testing.T) {
	f, err := os.CreateTemp(".", "concurrent_test_*.json")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	pipeline := New[*testutils.DummyRecord](Options{
		File: f,
	})
	defer pipeline.Close()
	require.NoError(t, pipeline.Open(context.Background()))

	var wg sync.WaitGroup
	numWorkers := 10
	itemsPerWorker := 100
	totalItems := numWorkers * itemsPerWorker

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < itemsPerWorker; j++ {
				record := &testutils.DummyRecord{
					Id:   fmt.Sprintf("%d-%d", i, j),
					Name: "worker",
				}
				_ = pipeline.ProcessItem(nil, record)
			}
		}()
	}

	wg.Wait()
	pipeline.Close()

	// Verify all items are present and valid JSON
	f, _ = os.Open(f.Name())
	defer f.Close()
	
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		var out testutils.DummyRecord
		err := json.Unmarshal(scanner.Bytes(), &out)
		assert.NoError(t, err, "Each line should be valid JSON")
		count++
	}
	assert.Equal(t, totalItems, count)
}

func TestExport2JSON_FlushMode(t *testing.T) {
	t.Run("ImmediateFlush", func(t *testing.T) {
		f, _ := os.CreateTemp(".", "flush_test_*.json")
		filename := f.Name()
		defer os.Remove(filename)

		pipeline := New[*testutils.DummyRecord](Options{
			File:      f,
			Immediate: true,
		})
		_ = pipeline.Open(context.Background())

		record := &testutils.DummyRecord{Id: "1", Name: "immediate"}
		_ = pipeline.ProcessItem(nil, record)

		info, _ := os.Stat(filename)
		assert.Greater(t, info.Size(), int64(0))
		pipeline.Close()
	})

	t.Run("BufferedFlush", func(t *testing.T) {
		f, _ := os.CreateTemp(".", "flush_test_*.json")
		filename := f.Name()
		defer os.Remove(filename)

		pipeline := New[*testutils.DummyRecord](Options{
			File:      f,
			Immediate: false,
		})
		_ = pipeline.Open(context.Background())

		record := &testutils.DummyRecord{Id: "1", Name: "buffered"}
		_ = pipeline.ProcessItem(nil, record)

		info, _ := os.Stat(filename)
		t.Logf("File size before close: %d", info.Size())

		pipeline.Close() // Closing pipeline triggers Flush()
		info, _ = os.Stat(filename)
		assert.Greater(t, info.Size(), int64(0))
	})
}
