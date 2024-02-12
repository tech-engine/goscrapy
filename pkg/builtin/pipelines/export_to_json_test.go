package pipelines

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExport2JSON(t *testing.T) {

	f, err := os.CreateTemp(".", "export_2_json")

	assert.NoError(t, err)

	defer os.Remove(f.Name())

	pipeline := Export2JSON[[]dummyRecord]()
	defer pipeline.Close()

	pipeline.WithWriteCloser(f)
	pipeline.WithImmediate()

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

	d := json.NewDecoder(f)

	var out = make([]dummyRecord, 2)
	err = d.Decode(&out)

	assert.NoError(t, err)

}
