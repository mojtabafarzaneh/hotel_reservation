package db

import (
	"context"
	"fmt"

	"github.com/mojtabafarzaneh/booking_room/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const userColl = "users"

type Dropper interface {
	Drop(context.Context) error
}

type UserStore interface {
	Dropper
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	GetUserByID(context.Context, string) (*types.User, error)
	GetUsers(context.Context) ([]*types.User, error)
	InsertUsers(context.Context, *types.User) (*types.User, error)
	DeleteUser(context.Context, string) error
	USerUpdate(ctx context.Context, filter bson.M, values types.UpdateUserParams) error
}
type MongoUserStore struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewMongoUserStore(client *mongo.Client) *MongoUserStore {
	return &MongoUserStore{
		client: client,
		coll:   client.Database(MongoDBNameEnvName).Collection(userColl),
	}
}

func (s *MongoUserStore) GetUsers(ctx context.Context) ([]*types.User, error) {
	cur, err := s.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var users []*types.User

	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *MongoUserStore) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	if err := s.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MongoUserStore) GetUserByID(ctx context.Context, id string) (*types.User, error) {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user types.User

	if err := s.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MongoUserStore) InsertUsers(ctx context.Context, user *types.User) (*types.User, error) {
	res, err := s.coll.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = res.InsertedID.(primitive.ObjectID)

	return user, nil

}

func (s *MongoUserStore) DeleteUser(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	//TODO: maybe it's a good idea to handle if we did not delete the user, like log it or smt
	_, err = s.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	return nil

}

func (s *MongoUserStore) USerUpdate(ctx context.Context, filter bson.M, values types.UpdateUserParams) error {
	update := bson.D{
		{
			Key: "$set", Value: values.Tobson(),
		},
	}

	_, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil

}

func (s *MongoUserStore) Drop(ctx context.Context) error {
	fmt.Println("---- dropping user collection")
	return s.coll.Drop(ctx)

}
