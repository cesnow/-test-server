package mongodb

import (
	"context"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/persistence/model"
	"kiyudesign.com/cesnow/light-server/pkg/stores/mon"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupAuthsTestModel(mt *mtest.T) (*AuthsModel, func()) {

	mon.Inject(mt.Name(), mt.Client)

	db := mon.MustNewMongo(mon.Config{
		URI:      mt.Name(),
		Database: mt.DB.Name(),
	})

	// Setup Redis
	s, err := miniredis.Run()
	assert.NoError(mt.T, err)

	cleanup := func() {
		s.Close()
	}

	cacheConf := cache.CacheConf{
		cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: s.Addr(),
				Type: "node",
			},
			Weight: 100,
		},
	}

	return NewTest(db, cacheConf), cleanup
}

func createTestAuthsDO() *model.Test {
	return &model.Test{
		ID:             primitive.NewObjectID(),
		AuthKeyId:      123,
		ApiId:          456,
		DeviceModel:    "test_device",
		SystemVersion:  "1.0",
		AppVersion:     "1.0",
		SystemLangCode: "en",
		LangPack:       "default",
		LangCode:       "en",
		Proxy:          "",
		Params:         "",
		ClientIp:       "127.0.0.1",
		ActiveAt:       time.Now().Unix(),
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
}

func TestAuthsModel_InsertOrUpdate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create test data
		testData := createTestAuthsDO()

		mt.AddMockResponses(
			// Response for FindOneAndUpdate (InsertOrUpdate)
			mtest.CreateSuccessResponse(bson.D{
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: testData.ID},
					{Key: "auth_key_id", Value: testData.AuthKeyId},
					{Key: "api_id", Value: testData.ApiId},
					{Key: "device_model", Value: testData.DeviceModel},
					{Key: "system_version", Value: testData.SystemVersion},
					{Key: "app_version", Value: testData.AppVersion},
					{Key: "system_lang_code", Value: testData.SystemLangCode},
					{Key: "lang_pack", Value: testData.LangPack},
					{Key: "lang_code", Value: testData.LangCode},
					{Key: "proxy", Value: testData.Proxy},
					{Key: "params", Value: testData.Params},
					{Key: "client_ip", Value: testData.ClientIp},
					{Key: "active_at", Value: testData.ActiveAt},
					{Key: "created_at", Value: testData.CreatedAt},
					{Key: "updated_at", Value: testData.UpdatedAt},
				}},
			}...),
		)

		// Setup model
		authsModel, cleanup := setupAuthsTestModel(mt)
		defer cleanup()

		// Test
		ctx := context.Background()
		updateId, err := authsModel.InsertOrUpdate(ctx, testData)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, updateId)
	})
}

func TestAuthsModel_SelectByAuthKeyId(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create test data
		testData := createTestAuthsDO()

		// Set up mock responses
		mt.AddMockResponses(
			// Response for FindOneAndUpdate (InsertOrUpdate)
			mtest.CreateSuccessResponse(bson.D{
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: testData.ID},
					{Key: "auth_key_id", Value: testData.AuthKeyId},
					{Key: "api_id", Value: testData.ApiId},
					{Key: "device_model", Value: testData.DeviceModel},
					{Key: "system_version", Value: testData.SystemVersion},
					{Key: "app_version", Value: testData.AppVersion},
					{Key: "system_lang_code", Value: testData.SystemLangCode},
					{Key: "lang_pack", Value: testData.LangPack},
					{Key: "lang_code", Value: testData.LangCode},
					{Key: "proxy", Value: testData.Proxy},
					{Key: "params", Value: testData.Params},
					{Key: "client_ip", Value: testData.ClientIp},
					{Key: "active_at", Value: testData.ActiveAt},
					{Key: "created_at", Value: testData.CreatedAt},
					{Key: "updated_at", Value: testData.UpdatedAt},
				}},
			}...),
			// Response for FindOne (SelectByAuthKeyId)
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "cursor", Value: bson.D{
					{Key: "id", Value: int64(0)},
					{Key: "ns", Value: "test_db.auths"},
					{Key: "firstBatch", Value: bson.A{
						bson.D{
							{Key: "_id", Value: testData.ID},
							{Key: "auth_key_id", Value: testData.AuthKeyId},
							{Key: "api_id", Value: testData.ApiId},
							{Key: "device_model", Value: testData.DeviceModel},
							{Key: "system_version", Value: testData.SystemVersion},
							{Key: "app_version", Value: testData.AppVersion},
							{Key: "system_lang_code", Value: testData.SystemLangCode},
							{Key: "lang_pack", Value: testData.LangPack},
							{Key: "lang_code", Value: testData.LangCode},
							{Key: "proxy", Value: testData.Proxy},
							{Key: "params", Value: testData.Params},
							{Key: "client_ip", Value: testData.ClientIp},
							{Key: "active_at", Value: testData.ActiveAt},
							{Key: "created_at", Value: testData.CreatedAt},
							{Key: "updated_at", Value: testData.UpdatedAt},
						},
					}},
				}},
			},
		)

		// Setup model
		authsModel, cleanup := setupAuthsTestModel(mt)
		defer cleanup()

		// Test
		ctx := context.Background()
		updateId, err := authsModel.InsertOrUpdate(ctx, testData)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, updateId)

		auth, err := authsModel.SelectByAuthKeyId(ctx, testData.AuthKeyId)
		assert.NoError(mt.T, err)
		assert.NotNil(mt.T, auth)
		assert.Equal(mt.T, testData.AuthKeyId, auth.AuthKeyId)
	})
}

func TestAuthsModel_SelectSessions(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create test data
		testData := createTestAuthsDO()

		// Set up mock responses
		mt.AddMockResponses(
			// Response for FindOneAndUpdate (InsertOrUpdate)
			mtest.CreateSuccessResponse(bson.D{
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: testData.ID},
					{Key: "auth_key_id", Value: testData.AuthKeyId},
					{Key: "api_id", Value: testData.ApiId},
					{Key: "device_model", Value: testData.DeviceModel},
					{Key: "system_version", Value: testData.SystemVersion},
					{Key: "app_version", Value: testData.AppVersion},
					{Key: "system_lang_code", Value: testData.SystemLangCode},
					{Key: "lang_pack", Value: testData.LangPack},
					{Key: "lang_code", Value: testData.LangCode},
					{Key: "proxy", Value: testData.Proxy},
					{Key: "params", Value: testData.Params},
					{Key: "client_ip", Value: testData.ClientIp},
					{Key: "active_at", Value: testData.ActiveAt},
					{Key: "created_at", Value: testData.CreatedAt},
					{Key: "updated_at", Value: testData.UpdatedAt},
				}},
			}...),
			// Response for Find (SelectSessions)
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "cursor", Value: bson.D{
					{Key: "id", Value: int64(0)},
					{Key: "ns", Value: "test_db.auths"},
					{Key: "firstBatch", Value: bson.A{
						bson.D{
							{Key: "_id", Value: testData.ID},
							{Key: "auth_key_id", Value: testData.AuthKeyId},
							{Key: "api_id", Value: testData.ApiId},
							{Key: "device_model", Value: testData.DeviceModel},
							{Key: "system_version", Value: testData.SystemVersion},
							{Key: "app_version", Value: testData.AppVersion},
							{Key: "system_lang_code", Value: testData.SystemLangCode},
							{Key: "lang_pack", Value: testData.LangPack},
							{Key: "lang_code", Value: testData.LangCode},
							{Key: "proxy", Value: testData.Proxy},
							{Key: "params", Value: testData.Params},
							{Key: "client_ip", Value: testData.ClientIp},
							{Key: "active_at", Value: testData.ActiveAt},
							{Key: "created_at", Value: testData.CreatedAt},
							{Key: "updated_at", Value: testData.UpdatedAt},
						},
					}},
				}},
			},
			mtest.CreateSuccessResponse(),
		)

		// Setup model
		authsModel, cleanup := setupAuthsTestModel(mt)
		defer cleanup()

		// Test
		ctx := context.Background()
		updateId, err := authsModel.InsertOrUpdate(ctx, testData)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, updateId)

		sessions, err := authsModel.SelectSessions(ctx, []int64{testData.AuthKeyId})
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, sessions)
		assert.Equal(mt.T, testData.AuthKeyId, sessions[0].AuthKeyId)
	})
}

func TestAuthsModel_SelectLangCode(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create test data
		testData := createTestAuthsDO()

		// Set up mock responses
		mt.AddMockResponses(
			// Response for FindOneAndUpdate (InsertOrUpdate)
			mtest.CreateSuccessResponse(bson.D{
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: testData.ID},
					{Key: "auth_key_id", Value: testData.AuthKeyId},
					{Key: "api_id", Value: testData.ApiId},
					{Key: "device_model", Value: testData.DeviceModel},
					{Key: "system_version", Value: testData.SystemVersion},
					{Key: "app_version", Value: testData.AppVersion},
					{Key: "system_lang_code", Value: testData.SystemLangCode},
					{Key: "lang_pack", Value: testData.LangPack},
					{Key: "lang_code", Value: testData.LangCode},
					{Key: "proxy", Value: testData.Proxy},
					{Key: "params", Value: testData.Params},
					{Key: "client_ip", Value: testData.ClientIp},
					{Key: "active_at", Value: testData.ActiveAt},
					{Key: "created_at", Value: testData.CreatedAt},
					{Key: "updated_at", Value: testData.UpdatedAt},
				}},
			}...),
			// Response for FindOne (SelectLangCode)
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "cursor", Value: bson.D{
					{Key: "id", Value: int64(0)},
					{Key: "ns", Value: "test_db.auths"},
					{Key: "firstBatch", Value: bson.A{
						bson.D{
							{Key: "_id", Value: testData.ID},
							{Key: "auth_key_id", Value: testData.AuthKeyId},
							{Key: "api_id", Value: testData.ApiId},
							{Key: "device_model", Value: testData.DeviceModel},
							{Key: "system_version", Value: testData.SystemVersion},
							{Key: "app_version", Value: testData.AppVersion},
							{Key: "system_lang_code", Value: testData.SystemLangCode},
							{Key: "lang_pack", Value: testData.LangPack},
							{Key: "lang_code", Value: testData.LangCode},
							{Key: "proxy", Value: testData.Proxy},
							{Key: "params", Value: testData.Params},
							{Key: "client_ip", Value: testData.ClientIp},
							{Key: "active_at", Value: testData.ActiveAt},
							{Key: "created_at", Value: testData.CreatedAt},
							{Key: "updated_at", Value: testData.UpdatedAt},
						},
					}},
				}},
			},
		)

		// Setup model
		authsModel, cleanup := setupAuthsTestModel(mt)
		defer cleanup()

		// Test
		ctx := context.Background()
		updateId, err := authsModel.InsertOrUpdate(ctx, testData)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, updateId)

		langCode, err := authsModel.SelectLangCode(ctx, testData.AuthKeyId)
		assert.NoError(mt.T, err)
		assert.Equal(mt.T, testData.LangCode, langCode)
	})
}

func TestAuthsModel_SelectLangPack(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create test data
		testData := createTestAuthsDO()

		// Set up mock responses
		mt.AddMockResponses(
			// Response for FindOneAndUpdate (InsertOrUpdate)
			mtest.CreateSuccessResponse(bson.D{
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: testData.ID},
					{Key: "auth_key_id", Value: testData.AuthKeyId},
					{Key: "api_id", Value: testData.ApiId},
					{Key: "device_model", Value: testData.DeviceModel},
					{Key: "system_version", Value: testData.SystemVersion},
					{Key: "app_version", Value: testData.AppVersion},
					{Key: "system_lang_code", Value: testData.SystemLangCode},
					{Key: "lang_pack", Value: testData.LangPack},
					{Key: "lang_code", Value: testData.LangCode},
					{Key: "proxy", Value: testData.Proxy},
					{Key: "params", Value: testData.Params},
					{Key: "client_ip", Value: testData.ClientIp},
					{Key: "active_at", Value: testData.ActiveAt},
					{Key: "created_at", Value: testData.CreatedAt},
					{Key: "updated_at", Value: testData.UpdatedAt},
				}},
			}...),
			// Response for FindOne (SelectLangPack)
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "cursor", Value: bson.D{
					{Key: "id", Value: int64(0)},
					{Key: "ns", Value: "test_db.auths"},
					{Key: "firstBatch", Value: bson.A{
						bson.D{
							{Key: "_id", Value: testData.ID},
							{Key: "auth_key_id", Value: testData.AuthKeyId},
							{Key: "api_id", Value: testData.ApiId},
							{Key: "device_model", Value: testData.DeviceModel},
							{Key: "system_version", Value: testData.SystemVersion},
							{Key: "app_version", Value: testData.AppVersion},
							{Key: "system_lang_code", Value: testData.SystemLangCode},
							{Key: "lang_pack", Value: testData.LangPack},
							{Key: "lang_code", Value: testData.LangCode},
							{Key: "proxy", Value: testData.Proxy},
							{Key: "params", Value: testData.Params},
							{Key: "client_ip", Value: testData.ClientIp},
							{Key: "active_at", Value: testData.ActiveAt},
							{Key: "created_at", Value: testData.CreatedAt},
							{Key: "updated_at", Value: testData.UpdatedAt},
						},
					}},
				}},
			},
		)

		// Setup model
		authsModel, cleanup := setupAuthsTestModel(mt)
		defer cleanup()

		// Test
		ctx := context.Background()
		updateId, err := authsModel.InsertOrUpdate(ctx, testData)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, updateId)

		langPack, err := authsModel.SelectLangPack(ctx, testData.AuthKeyId)
		assert.NoError(mt.T, err)
		assert.Equal(mt.T, testData.LangPack, langPack)
	})
}

func TestAuthsModel_SelectSessionsWithCB(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create test data
		testData := createTestAuthsDO()

		// Set up mock responses
		mt.AddMockResponses(
			// Response for FindOneAndUpdate (InsertOrUpdate)
			mtest.CreateSuccessResponse(bson.D{
				{Key: "value", Value: bson.D{
					{Key: "_id", Value: testData.ID},
					{Key: "auth_key_id", Value: testData.AuthKeyId},
					{Key: "api_id", Value: testData.ApiId},
					{Key: "device_model", Value: testData.DeviceModel},
					{Key: "system_version", Value: testData.SystemVersion},
					{Key: "app_version", Value: testData.AppVersion},
					{Key: "system_lang_code", Value: testData.SystemLangCode},
					{Key: "lang_pack", Value: testData.LangPack},
					{Key: "lang_code", Value: testData.LangCode},
					{Key: "proxy", Value: testData.Proxy},
					{Key: "params", Value: testData.Params},
					{Key: "client_ip", Value: testData.ClientIp},
					{Key: "active_at", Value: testData.ActiveAt},
					{Key: "created_at", Value: testData.CreatedAt},
					{Key: "updated_at", Value: testData.UpdatedAt},
				}},
			}...),
			// Response for Find (SelectSessionsWithCB)
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "cursor", Value: bson.D{
					{Key: "id", Value: int64(0)},
					{Key: "ns", Value: "test_db.auths"},
					{Key: "firstBatch", Value: bson.A{
						bson.D{
							{Key: "_id", Value: testData.ID},
							{Key: "auth_key_id", Value: testData.AuthKeyId},
							{Key: "api_id", Value: testData.ApiId},
							{Key: "device_model", Value: testData.DeviceModel},
							{Key: "system_version", Value: testData.SystemVersion},
							{Key: "app_version", Value: testData.AppVersion},
							{Key: "system_lang_code", Value: testData.SystemLangCode},
							{Key: "lang_pack", Value: testData.LangPack},
							{Key: "lang_code", Value: testData.LangCode},
							{Key: "proxy", Value: testData.Proxy},
							{Key: "params", Value: testData.Params},
							{Key: "client_ip", Value: testData.ClientIp},
							{Key: "active_at", Value: testData.ActiveAt},
							{Key: "created_at", Value: testData.CreatedAt},
							{Key: "updated_at", Value: testData.UpdatedAt},
						},
					}},
				}},
			},
		)

		// Setup model
		authsModel, cleanup := setupAuthsTestModel(mt)
		defer cleanup()

		// Test
		ctx := context.Background()
		updateId, err := authsModel.InsertOrUpdate(ctx, testData)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, updateId)

		var callbackCalled bool
		callback := func(sz, i int, v *model.Test) {
			callbackCalled = true
			assert.Equal(mt.T, testData.AuthKeyId, v.AuthKeyId)
		}

		sessions, err := authsModel.SelectSessionsWithCB(ctx, []int64{testData.AuthKeyId}, callback)
		assert.NoError(mt.T, err)
		assert.NotEmpty(mt.T, sessions)
		assert.True(mt.T, callbackCalled)
	})
}
