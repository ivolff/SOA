package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDbHandle struct {
	client   *mongo.Client
	mafiaDb  *mongo.Database
	comments *mongo.Collection
	sessions *mongo.Collection
}

func (dh *MongoDbHandle) InitConnection(username, password, host string, port int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbAddress := fmt.Sprintf("mongodb://%s:%s@%s:%d/", username, password, host, port)
	clientOptions := options.Client().ApplyURI(dbAddress)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	dh.client = client

	err = dh.client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	dh.mafiaDb = dh.client.Database("mafiaGraphQL")
	dh.sessions = dh.mafiaDb.Collection("sessions")
	dh.comments = dh.mafiaDb.Collection("comments")
	return nil
}
