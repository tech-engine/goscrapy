package pipelines

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/tech-engine/goscrapy/pkg/core"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
	"google.golang.org/api/option"
)

type export2FIREBASE[IN core.Job, OUT any, OR core.Output[IN, OUT]] struct {
	onOpenHook  OpenHook
	onCloseHook CloseHook
	ctx         context.Context
	ref         *db.Ref
}

func Export2FIREBASE[IN core.Job, OUT any](_url, filePath, collName string) (*export2FIREBASE[IN, OUT, core.Output[IN, OUT]], error) {
	ctx := context.Background()

	conf := &firebase.Config{
		DatabaseURL: _url,
	}

	opt := option.WithCredentialsFile(filePath)

	app, err := firebase.NewApp(ctx, conf, opt)

	if err != nil {
		return nil, fmt.Errorf("Export2FIREBASE: Error initializing app %w", err)
	}

	client, err := app.Database(ctx)

	if err != nil {
		return nil, fmt.Errorf("Export2FIREBASE: Error initializing Firebase client %w", err)
	}

	return &export2FIREBASE[IN, OUT, core.Output[IN, OUT]]{
		ctx: ctx,
		ref: client.NewRef(collName),
	}, nil
}

func (p *export2FIREBASE[IN, OUT, OR]) SetOpenHook(open OpenHook) *export2FIREBASE[IN, OUT, OR] {
	p.onOpenHook = open
	return p
}

func (p *export2FIREBASE[IN, OUT, OR]) SetCloseHook(close CloseHook) *export2FIREBASE[IN, OUT, OR] {
	p.onCloseHook = close
	return p
}

func (p *export2FIREBASE[IN, OUT, OR]) Open(ctx context.Context) error {
	if p.onOpenHook == nil {
		return nil
	}
	return p.onOpenHook(ctx)
}

func (p *export2FIREBASE[IN, OUT, OR]) Close() {
	if p.onCloseHook == nil {
		return
	}
	p.onCloseHook()
}

// your custom pipeline processing code goes here
func (p *export2FIREBASE[IN, OUT, OR]) ProcessItem(input any, original OR, MetaData metadata.MetaData) (any, error) {

	if original.IsEmpty() {
		return nil, nil
	}

	if _, err := p.ref.Push(p.ctx, original.Records()); err != nil {
		return nil, fmt.Errorf("Export2FIREBASE: error inserting data to DB %w", err)
	}

	return nil, nil
}
