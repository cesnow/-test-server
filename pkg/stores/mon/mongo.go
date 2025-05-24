package mon

import (
	"context"
	"github.com/zeromicro/go-zero/core/breaker"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
)

type Config struct {
	URI      string
	Database string
}

func MustNewMongo(config Config) *DB {
	client, err := getClient(config.URI)
	if err != nil {
		panic(err)
	}
	return &DB{
		config: &config,
		client: client,
	}
}

type DB struct {
	config *Config
	client *mongo.Client
}

func (d *DB) NewModel(collection string, opts ...Option) *Model {
	name := strings.Join([]string{d.config.URI, collection}, "/")
	brk := breaker.GetBreaker(d.config.URI)
	coll := newCollection(d.client.Database(d.config.Database).Collection(collection), brk)
	return newModel(name, d.client, coll, brk, opts...)
}

func (d *DB) Client() *mongo.Client {
	return d.client
}

func (d *DB) Close() error {
	return d.client.Disconnect(context.Background())
}

func (d *DB) GetCol(name string) *mongo.Collection {
	return d.client.Database(d.config.Database).Collection(name)
}

func (d *DB) GetDB(db string) *mongo.Database {
	return d.client.Database(db)
}

func (d *DB) GetDefaultDb() *mongo.Database {
	return d.client.Database(d.config.Database)
}
