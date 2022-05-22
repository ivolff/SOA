package db

import (
	"context"
	"log"

	"github.com/ivolff/graphql-mafia/graph/model"

	"go.mongodb.org/mongo-driver/bson"
)

func FailOnError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (dh *MongoDbHandle) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	insert, err := dh.sessions.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	var res model.Session
	err = dh.sessions.FindOne(ctx, bson.D{{"_id", insert.InsertedID}}).Decode(&res)
	return &res, err
}

func (dh *MongoDbHandle) GetAllSessions(ctx context.Context) ([]*model.Session, error) {
	res := make([]*model.Session, 0)

	iterator, err := dh.sessions.Find(ctx, bson.D{{}})
	if err != nil {
		return nil, err
	}

	for iterator.Next(context.TODO()) {
		var session model.Session
		err := iterator.Decode(&session)
		FailOnError("Failed to decode session iterator", err)

		comments, err := dh.GetSessionComments(ctx, session.ID)
		if err != nil {
			return nil, err
		}

		session.Comments = comments
		res = append(res, &session)
	}

	return res, err
}

func (dh *MongoDbHandle) GetSessionsByStatus(ctx context.Context, ongoing bool) ([]*model.Session, error) {
	res := make([]*model.Session, 0)

	iterator, err := dh.sessions.Find(ctx, bson.D{{"ongoing", ongoing}})
	if err != nil {
		return nil, err
	}

	for iterator.Next(context.TODO()) {
		var session model.Session
		err := iterator.Decode(&session)
		FailOnError("Failed to decode session iterator", err)

		comments, err := dh.GetSessionComments(ctx, session.ID)
		if err != nil {
			return nil, err
		}

		session.Comments = comments
		res = append(res, &session)
	}

	return res, err
}

func (dh *MongoDbHandle) GetSessionById(ctx context.Context, id string) (*model.Session, error) {
	res := model.Session{}

	err := dh.sessions.FindOne(ctx, bson.D{{"_id", id}}).Decode(&res)
	if err != nil {
		return nil, err
	}

	comments, err := dh.GetSessionComments(ctx, id)
	res.Comments = comments

	return &res, err
}

func (dh *MongoDbHandle) UpdateSessionById(ctx context.Context, id string, updated *model.Session) error {
	_, err := dh.sessions.ReplaceOne(ctx, bson.D{{"_id", id}}, updated)
	return err
}
