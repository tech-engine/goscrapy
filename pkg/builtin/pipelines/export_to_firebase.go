package pipelines

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"google.golang.org/api/option"
)

type export2FIREBASE[OUT any] struct {
	ctx context.Context
	ref *db.Ref
}

func Export2FIREBASE[OUT any](_url, filePath, collName string) (*export2FIREBASE[OUT], error) {
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

	return &export2FIREBASE[OUT]{
		ctx: ctx,
		ref: client.NewRef(collName),
	}, nil
}

func (p *export2FIREBASE[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2FIREBASE[OUT]) Close() {
}

// your custom pipeline processing code goes here
func (p *export2FIREBASE[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	if original.IsEmpty() {
		return nil
	}

	if _, err := p.ref.Push(p.ctx, original.Records()); err != nil {
		return fmt.Errorf("Export2FIREBASE: error inserting data to DB %w", err)
	}

	return nil
}
