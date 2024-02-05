package pipelines

import (
	"encoding/json"
	"os"

	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/net/context"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

type export2JSON[OUT any] struct {
	filename string
}

func Export2JSON[OUT any](args ...string) (*export2JSON[OUT], error) {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}
	return &export2JSON[OUT]{
		filename: filename,
	}, nil
}

func (p *export2JSON[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2JSON[OUT]) Close() {
}

func (p *export2JSON[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	if original.IsEmpty() {
		return nil
	}

	if p.filename == "" {
		p.filename = "JOB_" + original.Job().Id() + ".json"
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)

	if err != nil {
		return err
	}

	defer file.Close()

	jsonEncoder := json.NewEncoder(file)

	// Encode and write the JSON data
	if err := jsonEncoder.Encode(original.Records()); err != nil {
		return err
	}

	return nil
}
