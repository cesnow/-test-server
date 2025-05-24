package mongodb

import (
	"context"
	"errors"
	"fmt"
	"kiyudesign.com/cesnow/light-server/app/bff/configuration/internal/persistence/model"
	"kiyudesign.com/cesnow/light-server/pkg/stores/mon"
	"kiyudesign.com/cesnow/light-server/pkg/stores/monc"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthsModel struct {
	*monc.Model[model.Test]
}

func NewTest(db *mon.DB, cache cache.CacheConf) *AuthsModel {
	authModel, _ := monc.NewModel[model.Test](db, model.Test{}.TableName(), cache)
	return &AuthsModel{
		authModel,
	}
}

// InsertOrUpdate -
func (repo *AuthsModel) InsertOrUpdate(ctx context.Context, do *model.Test) (updatedId string, err error) {
	filter := bson.M{"auth_key_id": do.AuthKeyId}
	update := bson.M{
		"$set": bson.M{
			"auth_key_id":      do.AuthKeyId,
			"api_id":           do.ApiId,
			"device_model":     do.DeviceModel,
			"system_version":   do.SystemVersion,
			"app_version":      do.AppVersion,
			"system_lang_code": do.SystemLangCode,
			"lang_pack":        do.LangPack,
			"lang_code":        do.LangCode,
			"proxy":            do.Proxy,
			"params":           do.Params,
			"client_ip":        do.ClientIp,
			"active_at":        do.ActiveAt,
		},
	}

	opts := options.FindOneAndUpdate()
	opts.SetUpsert(true)
	opts.SetReturnDocument(options.After)

	var updatedDoc model.Test

	err = repo.Model.FindOneAndUpdateNoCache(ctx, &updatedDoc, filter, update, opts)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logx.WithContext(ctx).Errorf("FindOneAndUpdate in InsertOrUpdate(%v), error: %v", do, err)
		return "", err
	}

	updatedId = updatedDoc.ID.Hex()
	return
	// insert into auths(auth_key_id, api_id, device_model, system_version, app_version, system_lang_code, lang_pack, lang_code, proxy, params, client_ip, active_at) values (:auth_key_id, :api_id, :device_model, :system_version, :app_version, :system_lang_code, :lang_pack, :lang_code, :proxy, :params, :client_ip, :active_at) on duplicate key update api_id = values(api_id), device_model = values(device_model), system_version = values(system_version), app_version = values(app_version), system_lang_code = values(system_lang_code), lang_pack = values(lang_pack), lang_code = values(lang_code), proxy = values(proxy), params = values(params), client_ip = values(client_ip), active_at = values(active_at)
}

// InsertOrUpdateTx
func (repo *AuthsModel) InsertOrUpdateTx(ctx context.Context, session mongo.Session, do *model.Test) (updatedID string, err error) {
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		id, errTx := repo.InsertOrUpdate(sc, do)
		if errTx != nil {
			logx.WithContext(sc).Errorf("Transaction InsertOrUpdate failed: %v", errTx)
			return errTx
		}
		updatedID = id
		return nil
	})
	return
}

// SelectSessions -
func (repo *AuthsModel) SelectSessions(ctx context.Context, idList []int64) (rList []model.Test, err error) {
	var doc []model.Test
	filter := bson.M{"auth_key_id": bson.M{"$in": idList}}
	err = repo.Model.Find(ctx, &doc, filter)
	fmt.Printf("doc: %v\n", err)
	if err != nil {
		logx.WithContext(ctx).Errorf("Find in SelectSessions(%v), error: %v", idList, err)
		return nil, err
	}

	return doc, nil
	// select auth_key_id, api_id, device_model, system_version, app_version, system_lang_code, lang_pack, lang_code, client_ip, active_at from auths where auth_key_id in (:idList)
}

// SelectSessionsWithCB -
func (repo *AuthsModel) SelectSessionsWithCB(ctx context.Context, idList []int64, cb func(sz, i int, v *model.Test)) (rList []model.Test, err error) {
	rList, err = repo.SelectSessions(ctx, idList)
	if err != nil {
		return nil, err
	}

	sz := len(rList)
	if cb != nil {
		for i := 0; i < sz; i++ {
			cb(sz, i, &rList[i])
		}
	}

	return rList, nil
}

// SelectByAuthKeyId -
func (repo *AuthsModel) SelectByAuthKeyId(ctx context.Context, authKeyId int64) (rValue *model.Test, err error) {
	filter := bson.M{"auth_key_id": authKeyId, "deleted": false}

	auth := &model.Test{}
	err = repo.Model.FindOneNoCache(ctx, auth, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("FindOneNoCache in SelectByAuthKeyId(%v), error: %v", authKeyId, err)
		return nil, err
	}

	return auth, nil
	// select auth_key_id, api_id, device_model, system_version, app_version, system_lang_code, lang_pack, lang_code, client_ip, active_at from auths where auth_key_id = :auth_key_id and deleted = 0 limit 1
}

// SelectLangCode -
func (repo *AuthsModel) SelectLangCode(ctx context.Context, authKeyId int64) (rValue string, err error) {
	filter := bson.M{"auth_key_id": authKeyId}

	var result model.Test

	opts := options.FindOne()
	opts.SetProjection(bson.M{"lang_code": 1})

	err = repo.Model.FindOneNoCache(ctx, &result, filter, opts)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logx.WithContext(ctx).Errorf("FindOne in SelectLangCode(%d), error: %v", authKeyId, err)
		}
		return "", err
	}

	rValue = result.LangCode
	return
	//  select lang_code from auths where auth_key_id = :auth_key_id limit 1
}

// SelectLangPack -
func (repo *AuthsModel) SelectLangPack(ctx context.Context, authKeyId int64) (rValue string, err error) {
	filter := bson.M{"auth_key_id": authKeyId}

	var result model.Test

	opts := options.FindOne()
	opts.SetProjection(bson.M{"lang_pack": 1})

	err = repo.Model.FindOneNoCache(ctx, &result, filter, opts)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logx.WithContext(ctx).Errorf("FindOne in SelectLangPack(%d), error: %v", authKeyId, err)
		}
		return "", err
	}

	rValue = result.LangPack
	return
	// select lang_pack from auths where auth_key_id = :auth_key_id limit 1
}
