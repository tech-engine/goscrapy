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

type export2JSON[OUT any] struct {
	filename string
	file     io.WriteCloser
	buff     *bufio.Writer
}

func Export2JSON[OUT any]() *export2JSON[OUT] {
	return &export2JSON[OUT]{
		filename: fmt.Sprintf("JOB_%s.json", time.Now().UTC().Format("2006-01-02-15-04-05")),
	}
}

func (p *export2JSON[OUT]) WithWriteCloser(w io.WriteCloser) {
	p.file = w
}

func (p *export2JSON[OUT]) WithFilename(n string) {
	p.filename = n
}

func (p *export2JSON[OUT]) Open(ctx context.Context) error {
	if p.file != nil {
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

	if original.IsEmpty() {
		return nil
	}

	jsonEncoder := json.NewEncoder(p.buff)

	// Encode and write the JSON data
	if err := jsonEncoder.Encode(original.Records()); err != nil {
		return err
	}

	return nil
}
