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

// Export2CSV configuration struct.
// File Field will take precedence over Filename field.
type Export2CSVOpts struct {
	Filename string
	File     *os.File
}

type export2CSV[OUT any] struct {
	filename string
	file     *os.File
}

func Export2CSV[OUT any](opts ...Export2CSVOpts) *export2CSV[OUT] {
	e := &export2CSV[OUT]{
		filename: fmt.Sprintf("JOB_%s.csv", time.Now().UTC().Format("2006-01-02-15-04-05")),
	}

	if len(opts) > 0 {
		if opts[0].Filename != "" {
			e.filename = opts[0].Filename
		}

		if opts[0].File != nil {
			e.file = opts[0].File
		}
	}

	return e
}

func (p *export2CSV[OUT]) Open(ctx context.Context) error {
	if p.file != nil {
		p.filename = ""
		return nil
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)

	if err != nil {
		return err
	}

	p.file = file
	return nil
}

func (p *export2CSV[OUT]) Close() {
	p.file.Close()
}

func (p *export2CSV[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	fileInfo, err := p.file.Stat()

	if err != nil {
		return err
	}

	size := fileInfo.Size()

	data := []OUT{original.Record()}

	if size > 0 {
		err = gocsv.MarshalWithoutHeaders(data, p.file)
	} else {
		err = gocsv.MarshalFile(data, p.file)
	}

	return err
}
