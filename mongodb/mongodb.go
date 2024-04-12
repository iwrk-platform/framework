package mongodb

import (
	"context"
	"fmt"
	"github.com/kamva/mgm/v3"
	"github.com/uptrace/bun/migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Mongodb struct {
	Coll       func(m mgm.Model, opts ...*options.CollectionOptions) *mgm.Collection
	Config     *Config
	Logger     *zap.Logger
	Migrations *migrate.Migrations
}

func NewMongoDB(logger *zap.Logger, config *Config) *Mongodb {
	logger.Debug("mongodb://" + config.User + ":" + config.Password + "@" + config.Host + "/" + config.Database + "?retryWrites=true&replicaSet=dbrs&readPreference=primary&connectTimeoutMS=10000&authSource=" + config.Database + "&authMechanism=SCRAM-SHA-1")
	err := mgm.SetDefaultConfig(nil, config.Database, options.Client().ApplyURI("mongodb://"+config.User+":"+config.Password+"@"+config.Host+"/"+config.Database+"?retryWrites=true&replicaSet=dbrs&readPreference=primary&connectTimeoutMS=10000&authSource="+config.Database+"&authMechanism=SCRAM-SHA-1"))
	if err != nil {
		logger.Error(err.Error())
	}
	return &Mongodb{
		Config: config,
		Logger: logger,
		Coll:   mgm.Coll,
	}
}

type BaseModel struct {
	ID        int64 `json:"id" bson:"_id,omitempty"`
	CreatedAt int64 `json:"created_at" bson:"created_at"`
	UpdatedAt int64 `json:"updated_at" bson:"updated_at"`
	DeletedAt int64 `json:"deleted_at" bson:"deleted_at"`
}

func (b *BaseModel) Creating(collName string) error {
	b.CreatedAt = time.Now().Unix()
	b.UpdatedAt = time.Now().Unix()
	var counter struct {
		ID    string `bson:"_id"`
		Value int64  `bson:"value"`
	}
	res := mgm.CollectionByName("counter").FindOneAndUpdate(context.Background(),
		bson.D{{"_id", collName}},
		bson.M{"$inc": bson.M{"value": 1}},
		options.FindOneAndUpdate().SetUpsert(true),
		options.FindOneAndUpdate().SetReturnDocument(options.After))
	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to find one and update: %w", err)
	}
	if err := res.Decode(&counter); err != nil {
		return fmt.Errorf("failed to decode counter: %w", err)
	}
	b.SetID(counter.Value)
	return nil
}

func (b *BaseModel) SetID(id interface{}) {
	b.ID = id.(int64)
}

func (b *BaseModel) PrepareID(id interface{}) (interface{}, error) {
	if _, ok := id.(string); ok {
		n, err := strconv.ParseInt(id.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert id")
		}
		return n, nil
	}
	return id, nil
}
func (b *BaseModel) GetID() interface{} {
	return b.ID
}
