package pipelines

import (
	"context"
	"fmt"

	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type export2MONGODB[OUT any] struct {
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
}

func Export2MONGODB[OUT any](_url string, dbName string, collName string) (*export2MONGODB[OUT], error) {

	ctx := context.Background()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(_url).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		return nil, fmt.Errorf("Export2MONGODB: error connecting to DB %w", err)
	}

	var result bson.M

	if err := client.Database(dbName).RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		return nil, fmt.Errorf("Export2MONGODB: error connecting to DB %w", err)
	}

	collection := client.Database(dbName).Collection(collName)

	return &export2MONGODB[OUT]{
		ctx:        ctx,
		client:     client,
		collection: collection,
	}, nil
}

func (p *export2MONGODB[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2MONGODB[OUT]) Close() {
}

func (p *export2MONGODB[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	if original.IsEmpty() {
		return nil
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
		return fmt.Errorf("Export2MONGODB: error inserting data to DB %w", err)
	}

	return nil
}
