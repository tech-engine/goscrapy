package pipelines

import (
	"context"
	"fmt"

	"github.com/tech-engine/goscrapy/pkg/core"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type export2MONGODB[IN core.Job, OUT any, OR core.Output[IN, OUT]] struct {
	collection  *mongo.Collection
	onOpenHook  func(context.Context) error
	ctx         context.Context
	onCloseHook func()
}

func Export2MONGODB[IN core.Job, OUT any](_url string, dbName string, collName string) (*export2MONGODB[IN, OUT, core.Output[IN, OUT]], error) {

	ctx := context.Background()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(_url).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("Export2MONGODB: error connecting to DB %w", err)
	}

	collection := client.Database(dbName).Collection(collName)

	return &export2MONGODB[IN, OUT, core.Output[IN, OUT]]{
		collection: collection,
		ctx:        ctx,
	}, nil
}

func (p *export2MONGODB[IN, OUT, OR]) SetOpenHook(open OpenHook) *export2MONGODB[IN, OUT, OR] {
	p.onOpenHook = open
	return p
}

func (p *export2MONGODB[IN, OUT, OR]) SetCloseHook(close CloseHook) *export2MONGODB[IN, OUT, OR] {
	p.onCloseHook = close
	return p
}

func (p *export2MONGODB[IN, OUT, OR]) Open(ctx context.Context) error {
	if p.onOpenHook == nil {
		return nil
	}
	return p.onOpenHook(ctx)
}

func (p *export2MONGODB[IN, OUT, OR]) Close() {
	if p.onCloseHook == nil {
		return
	}
	p.onCloseHook()
}

func (p *export2MONGODB[IN, OUT, OR]) ProcessItem(input any, original OR, MetaData metadata.MetaData) (any, error) {

	if original.IsEmpty() {
		return nil, nil
	}

	documents := make([]any, 0, len(original.RecordsFlat()))

	for _, v := range original.RecordsFlat() {
		doc := primitive.D{}
		for i, key := range original.RecordKeys() {
			doc = append(doc, primitive.E{Key: key, Value: v[i]})
		}
		documents = append(documents, doc)
	}

	_, err := p.collection.InsertMany(p.ctx, documents)

	if err != nil {
		return nil, fmt.Errorf("Export2MONGODB: error inserting data to DB %w", err)
	}

	return nil, nil
}
