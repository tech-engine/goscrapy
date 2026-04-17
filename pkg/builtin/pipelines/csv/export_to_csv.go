package csv

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

type Options struct {
	Filename string
	File     *os.File
}

type export2CSV[OUT any] struct {
	filename  string
	file      *os.File
	mu        sync.Mutex
	fileEmpty bool
}

func New[OUT any](opts ...Options) *export2CSV[OUT] {
	p := &export2CSV[OUT]{
		filename: fmt.Sprintf("JOB_%s.csv", time.Now().UTC().Format("2006-01-02-15-04-05")),
	}

	if len(opts) > 0 {
		if opts[0].Filename != "" {
			p.filename = opts[0].Filename
		}

		if opts[0].File != nil {
			p.file = opts[0].File
		}
	}

	return p
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

	// pre-check if file is not empty
	info, err := file.Stat()
	if err == nil {
		p.fileEmpty = info.Size() > 0
	}

	return nil
}

func (p *export2CSV[OUT]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.file != nil {
		p.file.Close()
	}
}

func (p *export2CSV[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	data := []OUT{original.Record()}

	p.mu.Lock()
	defer p.mu.Unlock()

	var err error
	if p.fileEmpty {
		return gocsv.MarshalWithoutHeaders(data, p.file)
	}

	err = gocsv.MarshalFile(data, p.file)
	p.fileEmpty = true

	return err
}
