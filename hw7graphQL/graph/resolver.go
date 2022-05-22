package graph

import "github.com/ivolff/graphql-mafia/db"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DbHandle db.MongoDbHandle
}
