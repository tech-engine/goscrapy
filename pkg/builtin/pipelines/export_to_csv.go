package pipelines

import (
	"fmt"
	"os"
	"time"

	"context"

	"github.com/gocarina/gocsv"
	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

type export2CSV[OUT any] struct {
	filename string
	file     *os.File
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
	filename := p.filename
	if filename == "" {
		formattedTime := time.Now().UTC().Format("2023-07-27-00-00-00")
		filename = fmt.Sprintf("JOB_%s.csv", formattedTime)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)

	if err != nil {
		return err
	}

	p.file = file
	return err
}

func (p *export2CSV[OUT]) Close() {
	p.file.Close()
}

func (p *export2CSV[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	if original.IsEmpty() {
		return nil
	}

	fileInfo, err := os.Stat(p.filename)

	if err != nil {
		return err
	}

	size := fileInfo.Size()

	data := original.Records()

	if size > 0 {
		err = gocsv.MarshalWithoutHeaders(data, p.file)
	} else {
		err = gocsv.MarshalFile(data, p.file)
	}

	return err
}
