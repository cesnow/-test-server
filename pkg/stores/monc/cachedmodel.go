package monc

import (
	"context"
	"kiyudesign.com/cesnow/light-server/pkg/stores/mon"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// ErrNotFound is an alias of mongo.ErrNoDocuments.
	ErrNotFound = mongo.ErrNoDocuments

	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlight = syncx.NewSingleFlight()
	stats        = cache.NewStat("monc")
)

type Entity interface {
	TableName() string
}

// A Model is a mongo model that built with cache capability.
type Model[T Entity] struct {
	*mon.Model
	cache cache.Cache
}

// MustNewModel returns a Model with a cache cluster, exists on errors.
func MustNewModel[T Entity](db *mon.DB, collection string, c cache.CacheConf, opts ...cache.Option) *Model[T] {
	model, err := NewModel[T](db, collection, c, opts...)
	logx.Must(err)
	return model
}

// MustNewNodeModel returns a Model with a cache node, exists on errors.
func MustNewNodeModel[T Entity](db *mon.DB, collection string, rds *redis.Redis, opts ...cache.Option) *Model[T] {
	model, err := NewNodeModel[T](db, collection, rds, opts...)
	logx.Must(err)
	return model
}

// NewModel returns a Model with a cache cluster.
func NewModel[T Entity](db *mon.DB, collection string, conf cache.CacheConf, opts ...cache.Option) (*Model[T], error) {
	c := cache.New(conf, singleFlight, stats, mongo.ErrNoDocuments, opts...)
	return NewModelWithCache[T](db, collection, c)
}

// NewModelWithCache returns a Model with a custom cache.
func NewModelWithCache[T Entity](db *mon.DB, collection string, c cache.Cache) (*Model[T], error) {
	return newModel[T](db, collection, c)
}

// NewNodeModel returns a Model with a cache node.
func NewNodeModel[T Entity](db *mon.DB, collection string, rds *redis.Redis, opts ...cache.Option) (*Model[T], error) {
	c := cache.NewNode(rds, singleFlight, stats, mongo.ErrNoDocuments, opts...)
	return NewModelWithCache[T](db, collection, c)
}

// newModel returns a Model with the given cache.
func newModel[T Entity](db *mon.DB, collection string, c cache.Cache) (*Model[T], error) {
	model, err := mon.NewModel(db, collection)
	if err != nil {
		return nil, err
	}

	return &Model[T]{
		Model: model,
		cache: c,
	}, nil
}

// DelCache deletes the cache with given keys.
func (mm *Model[T]) DelCache(ctx context.Context, keys ...string) error {
	return mm.cache.DelCtx(ctx, keys...)
}

// DeleteOne deletes the document with given filter, and remove it from cache.
func (mm *Model[T]) DeleteOne(ctx context.Context, key string, filter any,
	opts ...*mopt.DeleteOptions) (int64, error) {
	val, err := mm.Model.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}

	if err := mm.DelCache(ctx, key); err != nil {
		return 0, err
	}

	return val, nil
}

// DeleteOneNoCache deletes the document with given filter.
func (mm *Model[T]) DeleteOneNoCache(ctx context.Context, filter any,
	opts ...*mopt.DeleteOptions) (int64, error) {
	return mm.Model.DeleteOne(ctx, filter, opts...)
}

// FindOne unmarshal a record into v with given key and query.
func (mm *Model[T]) FindOne(ctx context.Context, key string, v, filter any,
	opts ...*mopt.FindOneOptions) error {
	return mm.cache.TakeCtx(ctx, v, key, func(v any) error {
		return mm.Model.FindOne(ctx, v, filter, opts...)
	})
}

// FindOneNoCache unmarshal a record into v with query, without cache.
func (mm *Model[T]) FindOneNoCache(ctx context.Context, v, filter any,
	opts ...*mopt.FindOneOptions) error {
	return mm.Model.FindOne(ctx, v, filter, opts...)
}

// FindOneAndDelete deletes the document with given filter, and unmarshal it into v.
func (mm *Model[T]) FindOneAndDelete(ctx context.Context, key string, v, filter any,
	opts ...*mopt.FindOneAndDeleteOptions) error {
	if err := mm.Model.FindOneAndDelete(ctx, v, filter, opts...); err != nil {
		return err
	}

	return mm.DelCache(ctx, key)
}

// FindOneAndDeleteNoCache deletes the document with given filter, and unmarshal it into v.
func (mm *Model[T]) FindOneAndDeleteNoCache(ctx context.Context, v, filter any,
	opts ...*mopt.FindOneAndDeleteOptions) error {
	return mm.Model.FindOneAndDelete(ctx, v, filter, opts...)
}

// FindOneAndReplace replaces the document with given filter with replacement, and unmarshal it into v.
func (mm *Model[T]) FindOneAndReplace(ctx context.Context, key string, v T, filter any,
	replacement any, opts ...*mopt.FindOneAndReplaceOptions) error {
	if err := mm.Model.FindOneAndReplace(ctx, v, filter, replacement, opts...); err != nil {
		return err
	}

	return mm.DelCache(ctx, key)
}

// FindOneAndReplaceNoCache replaces the document with given filter with replacement, and unmarshal it into v.
func (mm *Model[T]) FindOneAndReplaceNoCache(ctx context.Context, v T, filter any,
	replacement any, opts ...*mopt.FindOneAndReplaceOptions) error {
	return mm.Model.FindOneAndReplace(ctx, v, filter, replacement, opts...)
}

// FindOneAndUpdate updates the document with given filter with update, and unmarshal it into v.
func (mm *Model[T]) FindOneAndUpdate(ctx context.Context, key string, v T, filter any,
	update any, opts ...*mopt.FindOneAndUpdateOptions) error {
	if err := mm.Model.FindOneAndUpdate(ctx, v, filter, update, opts...); err != nil {
		return err
	}

	return mm.DelCache(ctx, key)
}

// FindOneAndUpdateNoCache updates the document with given filter with update, and unmarshal it into v.
func (mm *Model[T]) FindOneAndUpdateNoCache(ctx context.Context, v *T, filter any,
	update any, opts ...*mopt.FindOneAndUpdateOptions) error {
	return mm.Model.FindOneAndUpdate(ctx, v, filter, update, opts...)
}

// GetCache unmarshal the cache into v with given key.
func (mm *Model[T]) GetCache(key string, v any) error {
	return mm.cache.Get(key, v)
}

// InsertOne inserts a single document into the collection, and remove the cache placeholder.
func (mm *Model[T]) InsertOne(ctx context.Context, key string, document T,
	opts ...*mopt.InsertOneOptions) (*mongo.InsertOneResult, error) {
	res, err := mm.Model.InsertOne(ctx, document, opts...)
	if err != nil {
		return nil, err
	}

	if err = mm.DelCache(ctx, key); err != nil {
		return nil, err
	}

	return res, nil
}

func (mm *Model[T]) Test(ctx context.Context, doc T) error {
	return nil
}

// InsertOneNoCache inserts a single document into the collection.
func (mm *Model[T]) InsertOneNoCache(ctx context.Context, document *T,
	opts ...*mopt.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return mm.Model.InsertOne(ctx, document, opts...)
}

// ReplaceOne replaces a single document in the collection, and remove the cache.
func (mm *Model[T]) ReplaceOne(ctx context.Context, key string, filter, replacement any,
	opts ...*mopt.ReplaceOptions) (*mongo.UpdateResult, error) {
	res, err := mm.Model.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		return nil, err
	}

	if err = mm.DelCache(ctx, key); err != nil {
		return nil, err
	}

	return res, nil
}

// ReplaceOneNoCache replaces a single document in the collection.
func (mm *Model[T]) ReplaceOneNoCache(ctx context.Context, filter, replacement any,
	opts ...*mopt.ReplaceOptions) (*mongo.UpdateResult, error) {
	return mm.Model.ReplaceOne(ctx, filter, replacement, opts...)
}

// SetCache sets the cache with given key and value.
func (mm *Model[T]) SetCache(key string, v any) error {
	return mm.cache.Set(key, v)
}

// UpdateByID updates the document with given id with update, and remove the cache.
func (mm *Model[T]) UpdateByID(ctx context.Context, key string, id, update any,
	opts ...*mopt.UpdateOptions) (*mongo.UpdateResult, error) {
	res, err := mm.Model.UpdateByID(ctx, id, update, opts...)
	if err != nil {
		return nil, err
	}

	if err = mm.DelCache(ctx, key); err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateByIDNoCache updates the document with given id with update.
func (mm *Model[T]) UpdateByIDNoCache(ctx context.Context, id, update any,
	opts ...*mopt.UpdateOptions) (*mongo.UpdateResult, error) {
	return mm.Model.UpdateByID(ctx, id, update, opts...)
}

// UpdateMany updates the documents that match filter with update, and remove the cache.
func (mm *Model[T]) UpdateMany(ctx context.Context, keys []string, filter, update any,
	opts ...*mopt.UpdateOptions) (*mongo.UpdateResult, error) {
	res, err := mm.Model.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, err
	}

	if err = mm.DelCache(ctx, keys...); err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateManyNoCache updates the documents that match filter with update.
func (mm *Model[T]) UpdateManyNoCache(ctx context.Context, filter, update any,
	opts ...*mopt.UpdateOptions) (*mongo.UpdateResult, error) {
	return mm.Model.UpdateMany(ctx, filter, update, opts...)
}

// UpdateOne updates the first document that matches filter with update, and remove the cache.
func (mm *Model[T]) UpdateOne(ctx context.Context, key string, filter any, update T,
	opts ...*mopt.UpdateOptions) (*mongo.UpdateResult, error) {
	res, err := mm.Model.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, err
	}

	if err = mm.DelCache(ctx, key); err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateOneNoCache updates the first document that matches filter with update.
func (mm *Model[T]) UpdateOneNoCache(ctx context.Context, filter, update any,
	opts ...*mopt.UpdateOptions) (*mongo.UpdateResult, error) {
	return mm.Model.UpdateOne(ctx, filter, update, opts...)
}

type (
	ExecFn      func(ctx context.Context, model *mon.Model) (string, error)
	QueryFn     func(ctx context.Context, model *mon.Model, v any) error
	KeysQueryFn func(ctx context.Context, keys ...string) (map[string]any, error)
	CacheFn     func(k, v string) (any, error)
)

// ExecCtx runs given exec on given keys, and returns execution result.
func (mm *Model[T]) ExecCtx(ctx context.Context, exec ExecFn, keys ...string) (lastInsertId string, err error) {
	lastInsertId, err = exec(ctx, mm.Model)
	if err != nil {
		return
	}

	err = mm.DelCache(ctx, keys...)
	return
}

// QueryCtx unmarshal into v with given key and query func.
func (mm *Model[T]) QueryCtx(ctx context.Context, v any, key string, query QueryFn) error {
	return mm.cache.TakeCtx(ctx, v, key, func(v any) error {
		return query(ctx, mm.Model, v)
	})
}

// FindCustom unmarshal into v with given key and query func.
//func (mm *Model[T) FindCustom(ctx context.Context, query KeysQueryFn, calVal CacheFn, keys ...string) error {
//	return mm.cache.TakesCtx(
//		ctx,
//		func(keys ...string) (map[string]any, error) {
//			return query(ctx, keys...)
//		},
//		calVal,
//		keys...)
//}
