package pipelines

import (
	"os"

	"context"

	"github.com/gocarina/gocsv"
	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

type export2CSV[OUT any] struct {
	filename string
}

func Export2CSV[OUT any](args ...string) *export2CSV[OUT] {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}
	return &export2CSV[OUT]{
		filename: filename,
	}
}

func (p *export2CSV[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2CSV[OUT]) Close() {
}

func (p *export2CSV[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	if original.IsEmpty() {
		return nil
	}

	if p.filename == "" {
		p.filename = "JOB_" + original.Job().Id() + ".csv"
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)

	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := os.Stat(p.filename)

	if err != nil {
		return err
	}

	size := fileInfo.Size()

	data := original.Records()

	if size > 0 {
		err = gocsv.MarshalWithoutHeaders(data, file)
	} else {
		err = gocsv.MarshalFile(data, file)
	}

	return err
}
