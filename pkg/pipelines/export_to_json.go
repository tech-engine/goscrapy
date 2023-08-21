package pipelines

import (
	"encoding/json"
	"os"

	"github.com/tech-engine/goscrapy/pkg/core"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
	"golang.org/x/net/context"
)

type export2JSON[IN core.Job, OUT any, OR core.Output[IN, OUT]] struct {
	filename    string
	onOpenHook  func(context.Context) error
	onCloseHook func()
}

func Export2JSON[IN core.Job, OUT any](args ...string) (*export2JSON[IN, OUT, core.Output[IN, OUT]], error) {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}
	return &export2JSON[IN, OUT, core.Output[IN, OUT]]{
		filename: filename,
	}, nil
}

func (p *export2JSON[IN, OUT, OR]) SetOpenHook(open OpenHook) *export2JSON[IN, OUT, OR] {
	p.onOpenHook = open
	return p
}

func (p *export2JSON[IN, OUT, OR]) SetCloseHook(close CloseHook) *export2JSON[IN, OUT, OR] {
	p.onCloseHook = close
	return p
}

func (p *export2JSON[IN, OUT, OR]) Open(ctx context.Context) error {
	if p.onOpenHook == nil {
		return nil
	}
	return p.onOpenHook(ctx)
}

func (p *export2JSON[IN, OUT, OR]) Close() {
	if p.onCloseHook == nil {
		return
	}
	p.onCloseHook()
}

func (p *export2JSON[IN, OUT, OR]) ProcessItem(input any, original OR, metadata metadata.MetaData) (any, error) {

	if original.IsEmpty() {
		return nil, nil
	}

	if p.filename == "" {
		p.filename = "JOB_" + original.Job().Id() + ".json"
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0640)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	jsonEncoder := json.NewEncoder(file)

	// Encode and write the JSON data
	if err := jsonEncoder.Encode(original.Records()); err != nil {
		return nil, err
	}

	return nil, nil
}
