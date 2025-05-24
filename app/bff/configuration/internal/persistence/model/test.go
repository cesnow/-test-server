package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Test struct {
	AuthKeyId      int64              `bson:"auth_key_id" db:"auth_key_id" json:"auth_key_id"`
	ApiId          int32              `bson:"api_id" db:"api_id" json:"api_id"`
	DeviceModel    string             `bson:"device_model" db:"device_model" json:"device_model"`
	SystemVersion  string             `bson:"system_version" db:"system_version" json:"system_version"`
	AppVersion     string             `bson:"app_version" db:"app_version" json:"app_version"`
	SystemLangCode string             `bson:"system_lang_code" db:"system_lang_code" json:"system_lang_code"`
	LangPack       string             `bson:"lang_pack" db:"lang_pack" json:"lang_pack"`
	LangCode       string             `bson:"lang_code" db:"lang_code" json:"lang_code"`
	SystemCode     string             `bson:"system_code" db:"system_code" json:"system_code"`
	Proxy          string             `bson:"proxy" db:"proxy" json:"proxy"`
	Params         string             `bson:"params" db:"params" json:"params"`
	ClientIp       string             `bson:"client_ip" db:"client_ip" json:"client_ip"`
	ActiveAt       int64              `bson:"active_at" db:"active_at" json:"active_at"`
	Deleted        bool               `bson:"deleted" db:"deleted" json:"deleted"`
	CreatedAt      int64              `bson:"created_at" db:"created_at" json:"created_at"`
	UpdatedAt      int64              `bson:"updated_at" db:"updated_at" json:"updated_at"`
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
}

func (Test) TableName() string {
	return "auths"
}
