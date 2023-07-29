package pipelines

import (
	"os"
	"strconv"
	"time"

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

func Export2CSV[IN core.Job, OUT any](args ...string) *export2CSV[IN, OUT, core.Output[IN, OUT]] {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}
	return &export2CSV[IN, OUT, core.Output[IN, OUT]]{
		filename: filename,
	}
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

	filename := MetaData.Get("JOB_ID").(string) + "_" + strconv.FormatInt(time.Now().UnixMicro(), 10) + "_" + MetaData.Get("JOB_NAME").(string) + ".csv"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, err := os.Stat(filename)

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
