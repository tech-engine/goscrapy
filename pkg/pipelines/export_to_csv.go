package pipelines

import (
	"os"

	"context"

	"github.com/gocarina/gocsv"
	"github.com/tech-engine/goscrapy/pkg/core"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

type export2CSV[IN core.Job, OUT any, OR core.Output[IN, OUT]] struct {
	filename    string
	onOpenHook  OpenHook
	onCloseHook CloseHook
}

func Export2CSV[IN core.Job, OUT any](args ...string) (*export2CSV[IN, OUT, core.Output[IN, OUT]], error) {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}
	return &export2CSV[IN, OUT, core.Output[IN, OUT]]{
		filename: filename,
	}, nil
}

func (p *export2CSV[IN, OUT, OR]) SetOpenHook(open OpenHook) *export2CSV[IN, OUT, OR] {
	p.onOpenHook = open
	return p
}

func (p *export2CSV[IN, OUT, OR]) SetCloseHook(close CloseHook) *export2CSV[IN, OUT, OR] {
	p.onCloseHook = close
	return p
}

func (p *export2CSV[IN, OUT, OR]) Open(ctx context.Context) error {
	if p.onOpenHook == nil {
		return nil
	}
	return p.onOpenHook(ctx)
}

func (p *export2CSV[IN, OUT, OR]) Close() {
	if p.onCloseHook == nil {
		return
	}
	p.onCloseHook()
}

func (p *export2CSV[IN, OUT, OR]) ProcessItem(input any, original OR, MetaData metadata.MetaData) (any, error) {

	if original.IsEmpty() {
		return nil, nil
	}

	if p.filename == "" {
		p.filename = "JOB_" + original.Job().Id() + ".csv"
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, err := os.Stat(p.filename)

	if err != nil {
		return nil, err
	}

	size := fileInfo.Size()

	data := original.Records()

	if size > 0 {
		err = gocsv.MarshalWithoutHeaders(data, file)
	} else {
		err = gocsv.MarshalFile(data, file)
	}

	return nil, err
}
