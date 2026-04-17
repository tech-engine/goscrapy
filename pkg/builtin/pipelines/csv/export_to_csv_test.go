package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport2CSV(t *testing.T) {
	f, err := os.CreateTemp(".", "export_test_*.csv")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	pipeline := New[*testutils.DummyRecord](Options{
		File: f,
	})
	defer pipeline.Close()

	require.NoError(t, pipeline.Open(context.Background()))

	record := &testutils.DummyRecord{Id: "1", Name: "rick"}
	assert.NoError(t, pipeline.ProcessItem(nil, record))

	f.Seek(0, 0)
	reader := csv.NewReader(f)
	csvRecords, err := reader.ReadAll()
	assert.NoError(t, err)

	assert.Len(t, csvRecords, 2) // Header + 1 Row
	assert.Equal(t, []string{"id", "name"}, csvRecords[0])
	assert.Equal(t, []string{"1", "rick"}, csvRecords[1])
}

func TestExport2CSV_Concurrent(t *testing.T) {
	f, err := os.CreateTemp(".", "concurrent_test_*.csv")
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

	// Verify results
	f.Seek(0, 0)
	reader := csv.NewReader(f)
	csvRecords, err := reader.ReadAll()
	assert.NoError(t, err)

	// Header + totalItems
	assert.Equal(t, totalItems+1, len(csvRecords), "Total rows should be items + 1 header")

	// Ensure no duplicate header
	headerCount := 0
	for _, row := range csvRecords {
		if row[0] == "id" && row[1] == "name" {
			headerCount++
		}
	}
	assert.Equal(t, 1, headerCount, "Should only have one header even with concurrent writes")
}

func TestExport2CSV_AppendMode(t *testing.T) {
	tmpFile, err := os.CreateTemp(".", "append_test_*.csv")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// 1. Pre-populate file with a header and one row
	_, _ = tmpFile.WriteString("id,name\n0,existing\n")
	tmpFile.Close()

	// 2. Open pipeline in append mode
	pipeline := New[*testutils.DummyRecord](Options{
		Filename: tmpPath,
	})
	require.NoError(t, pipeline.Open(context.Background()))
	defer pipeline.Close()

	// 3. Process new item
	record := &testutils.DummyRecord{Id: "1", Name: "new"}
	assert.NoError(t, pipeline.ProcessItem(nil, record))
	pipeline.Close()

	// 4. Verify no duplicate header
	f, _ := os.Open(tmpPath)
	defer f.Close()
	reader := csv.NewReader(f)
	csvRecords, _ := reader.ReadAll()

	assert.Len(t, csvRecords, 3) // Header + Existing + New
	assert.Equal(t, []string{"id", "name"}, csvRecords[0])
	assert.Equal(t, []string{"0", "existing"}, csvRecords[1])
	assert.Equal(t, []string{"1", "new"}, csvRecords[2])
}
