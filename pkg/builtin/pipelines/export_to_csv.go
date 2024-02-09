package pipelines

import (
	"fmt"
	"io"
	"os"
	"time"

	"context"

	"github.com/gocarina/gocsv"
	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

type export2CSV[OUT any] struct {
	filename string
	file     io.WriteCloser
}

func Export2CSV[OUT any]() *export2CSV[OUT] {
	return &export2CSV[OUT]{
		filename: fmt.Sprintf("JOB_%s.csv", time.Now().UTC().Format("2023-07-27-00-00-00")),
	}
}

func (p *export2CSV[OUT]) WithWriteCloser(w io.WriteCloser) *export2CSV[OUT] {
	p.file = w
	return p
}

func (p *export2CSV[OUT]) WithFilename(n string) *export2CSV[OUT] {
	p.filename = n
	return p
}

func (p *export2CSV[OUT]) Open(ctx context.Context) error {

	if p.file != nil {
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
		err = gocsv.MarshalFile(data, p.file.(*os.File))
	}

	return err
}
