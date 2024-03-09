package pipelines

import (
	"context"
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExport2CSV(t *testing.T) {

	f, err := os.CreateTemp(".", "export_2_csv.csv")

	assert.NoError(t, err)

	defer os.Remove(f.Name())

	pipeline := Export2CSV[*dummyRecord](Export2CSVOpts{
		File: f,
	})

	defer pipeline.Close()

	err = pipeline.Open(context.Background())

	assert.NoError(t, err)

	record := &dummyRecord{Id: "1", Name: "rick"}

	err = pipeline.ProcessItem(nil, record)

	assert.NoError(t, err)

	f.Seek(0, 0)

	reader := csv.NewReader(f)

	csvRecords, err := reader.ReadAll()

	assert.NoError(t, err)

	assert.Equal(t, convertToSliceOfStrings(record), csvRecords[1])

}

func convertToSliceOfStrings(record *dummyRecord) []string {
	return []string{record.Id, record.Name}
}
