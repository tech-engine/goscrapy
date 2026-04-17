package json

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

// Options configuration struct.
// Immediate=true, flushes bufio.Writer immediately after processing.
type Options struct {
	Filename  string
	File      io.WriteCloser
	Immediate bool
}

type export2JSON[OUT any] struct {
	filename       string
	file           io.WriteCloser
	buff           *bufio.Writer
	immediateFlush bool
	mu             sync.Mutex
}

func New[OUT any](opts ...Options) *export2JSON[OUT] {
	p := &export2JSON[OUT]{
		filename: fmt.Sprintf("JOB_%s.json", time.Now().UTC().Format("2006-01-02-15-04-05")),
	}

	if len(opts) > 0 {
		opt := opts[0]

		if opt.Filename != "" {
			p.filename = opt.Filename
		}

		if opt.File != nil {
			p.file = opt.File
		}

		p.immediateFlush = opt.Immediate
	}

	return p
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
	p.mu.Lock()
	defer p.mu.Unlock()

	p.buff.Flush()
	p.file.Close()
}

func (p *export2JSON[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	jsonEncoder := json.NewEncoder(p.buff)
	record := original.Record()

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := jsonEncoder.Encode(record); err != nil {
		return err
	}

	if p.immediateFlush {
		return p.buff.Flush()
	}

	return nil
}
