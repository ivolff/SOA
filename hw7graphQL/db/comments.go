package db

import (
	"context"
	"errors"

	"github.com/ivolff/graphql-mafia/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (dbh *MongoDbHandle) AddComment(ctx context.Context, comment *model.Comment) (*mongo.InsertOneResult, error) {
	if asscSession, err := dbh.GetSessionById(ctx, comment.SessionID); err != nil || asscSession == nil {
		return nil, errors.New("no session found")
	}

	res, err := dbh.comments.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (dbh *MongoDbHandle) GetSessionComments(ctx context.Context, sessionId string) ([]*model.Comment, error) {
	var res []*model.Comment

	iterator, err := dbh.comments.Find(ctx, bson.D{{"sessionId", sessionId}})
	if err != nil {
		return nil, err
	}

	if err = iterator.All(ctx, &res); err != nil {
		return nil, err
	}

	return res, err
}
