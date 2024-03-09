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

	pipeline := Export2JSON[*dummyRecord](Export2JSONOpts{
		File:      f,
		Immediate: true,
	})

	defer pipeline.Close()

	err = pipeline.Open(context.Background())

	assert.NoError(t, err)

	record := &dummyRecord{Id: "1", Name: "rick"}

	err = pipeline.ProcessItem(nil, record)

	assert.NoError(t, err)

	f.Seek(0, 0)

	d := json.NewDecoder(f)

	var out dummyRecord
	err = d.Decode(&out)

	assert.NoError(t, err)

}
