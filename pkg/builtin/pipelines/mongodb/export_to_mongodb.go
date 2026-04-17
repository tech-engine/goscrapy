package mongodb

import (
	"context"
	"fmt"
	"log"

	"github.com/tech-engine/goscrapy/pkg/core"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type export2MongoDB[OUT any] struct {
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
}

func New[OUT any](url string, dbName string, collName string) *export2MongoDB[OUT] {

	ctx := context.Background()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(url).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		log.Printf("Export2MONGODB: error connecting to DB %s", err)
		return nil
	}

	var result bson.M

	if err := client.Database(dbName).RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		log.Printf("Export2MONGODB: error connecting to DB %s", err)
		return nil
	}

	collection := client.Database(dbName).Collection(collName)

	return &export2MongoDB[OUT]{
		ctx:        ctx,
		client:     client,
		collection: collection,
	}
}

func (p *export2MongoDB[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *export2MongoDB[OUT]) Close() {
	if p.client != nil {
		p.client.Disconnect(p.ctx)
	}
}

func (p *export2MongoDB[OUT]) ProcessItem(item pm.IPipelineItem, original core.IOutput[OUT]) error {

	doc := primitive.D{}
	recordFlat := original.RecordFlat()

	for i, key := range original.RecordKeys() {
		doc = append(doc, primitive.E{Key: key, Value: recordFlat[i]})
	}

	_, err := p.collection.InsertMany(p.ctx, []any{doc})

	if err != nil {
		return fmt.Errorf("Export2MONGODB: error inserting data to DB %w", err)
	}

	return nil
}
