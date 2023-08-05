package pipelines

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
	"golang.org/x/net/context"
)

type export2JSON[IN core.Job, OUT any, OR core.Output[IN, OUT]] struct {
	filename    string
	onOpenHook  func(context.Context) error
	onCloseHook func()
}

func Export2JSON[IN core.Job, OUT any](args ...string) *export2JSON[IN, OUT, core.Output[IN, OUT]] {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}
	return &export2JSON[IN, OUT, core.Output[IN, OUT]]{
		filename: filename,
	}
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

	byteData, err := json.Marshal(original.Records())

	if err != nil {
		return nil, err
	}

	if p.filename == "" {
		p.filename = "JOB_" + original.Job().Id() + "_" + strconv.FormatInt(time.Now().UnixMicro(), 10) + ".json"
	}

	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0640)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	_, err = file.Write(byteData)

	return nil, err
}
