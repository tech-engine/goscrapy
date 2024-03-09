package pipelines

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"golang.org/x/net/context"
)

// Immediate: export2JSON internally creates a bufio.Writer from provided io.Writer.
// Immediate=true, flushes bufio.Writer immediately after processing.
type Export2JSONOpts struct {
	Filename  string
	File      io.WriteCloser
	Immediate bool
}

type export2JSON[OUT any] struct {
	filename       string
	file           io.WriteCloser
	buff           *bufio.Writer
	immediateFlush bool
}

func Export2JSON[OUT any](opts ...Export2JSONOpts) *export2JSON[OUT] {
	e := &export2JSON[OUT]{
		filename: fmt.Sprintf("JOB_%s.json", time.Now().UTC().Format("2006-01-02-15-04-05")),
	}

	if len(opts) > 0 {
		opt := opts[0]

		if opt.Filename != "" {
			e.filename = opt.Filename
		}

		if opt.File != nil {
			e.file = opt.File
		}

		e.immediateFlush = opt.Immediate
	}

	return e
}

func (p *export2JSON[OUT]) Open(ctx context.Context) error {
	if p.file != nil {
		p.buff = bufio.NewWriter(p.file)
		return nil
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)

	if err != nil {
		return err
	}

	p.file = file
	p.buff = bufio.NewWriter(p.file)
	return nil
}

func (p *export2JSON[OUT]) Close() {
	// flushed data to writer
	p.buff.Flush()
	p.file.Close()
}

func (p *export2JSON[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	jsonEncoder := json.NewEncoder(p.buff)

	// Encode and write the JSON data
	if err := jsonEncoder.Encode(original.Record()); err != nil {
		return err
	}

	if p.immediateFlush {
		p.buff.Flush()
	}

	return nil
}
