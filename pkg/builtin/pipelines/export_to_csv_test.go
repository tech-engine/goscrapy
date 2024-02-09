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

	pipeline := Export2CSV[[]dummyRecord]().WithFile(f)
	defer pipeline.Close()

	err = pipeline.Open(context.Background())

	assert.NoError(t, err)

	records := []dummyRecord{
		{Id: "1", Name: "rick"},
		{Id: "2", Name: "morty"},
	}

	err = pipeline.ProcessItem(nil, &dummyOutput{
		records: records,
	})

	assert.NoError(t, err)

	f.Seek(0, 0)

	reader := csv.NewReader(f)

	csvRecords, err := reader.ReadAll()

	assert.NoError(t, err)

	csvRecords = csvRecords[1:]

	assert.Equal(t, convertToSliceOfStrings(records), csvRecords)

}

func convertToSliceOfStrings(records []dummyRecord) [][]string {
	result := make([][]string, len(records))

	for i, record := range records {
		result[i] = []string{record.Id, record.Name}
	}

	return result
}
