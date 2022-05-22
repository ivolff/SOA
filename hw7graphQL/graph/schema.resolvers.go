package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/ivolff/graphql-mafia/graph/generated"
	"github.com/ivolff/graphql-mafia/graph/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *mutationResolver) StartSession(ctx context.Context, input model.NewSession) (*model.Session, error) {
	session := model.Session{
		ID:      primitive.NewObjectID().Hex(),
		Name:    input.Name,
		Ongoing: true,
		Players: []string{input.Host},
	}

	return r.DbHandle.CreateSession(ctx, &session)
}

func (r *mutationResolver) AddPlayer(ctx context.Context, input model.NewPlayer) (*model.Session, error) {
	session, err := r.DbHandle.GetSessionById(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}

	if !session.Ongoing {
		return nil, errors.New("session has already been terminated, you can't add new participants")
	}

	session.Players = append(session.Players, input.UserID)
	err = r.DbHandle.UpdateSessionById(ctx, input.SessionID, session)
	if err != nil {
		return nil, err
	}

	return session, err
}

func (r *mutationResolver) AddComment(ctx context.Context, input model.NewComment) (string, error) {
	comment := model.Comment(input)

	_, err := r.DbHandle.AddComment(ctx, &comment)
	if err != nil {
		return "", err
	}

	return "Successfully added comment", err
}

func (r *mutationResolver) EndSession(ctx context.Context, sessionID string, timestamp *int) (string, error) {
	session, err := r.DbHandle.GetSessionById(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if !session.Ongoing {
		return "", errors.New("session has already been terminated")
	}

	session.Ongoing = false
	session.Timestamp = *timestamp
	err = r.DbHandle.UpdateSessionById(ctx, sessionID, session)
	if err != nil {
		return "", err
	}

	return "Session terminated", err
}

func (r *queryResolver) Sessions(ctx context.Context, ongoing *bool, sessionID *string) ([]*model.Session, error) {
	if sessionID != nil {
		res, err := r.DbHandle.GetSessionById(ctx, *sessionID)
		return []*model.Session{res}, err
	}

	return r.DbHandle.GetSessionsByStatus(ctx, *ongoing)
}

func (r *queryResolver) ScoreBoard(ctx context.Context, sessionID *string) (*model.Board, error) {
	res, err := r.DbHandle.GetSessionById(ctx, *sessionID)
	board := model.Board{}
	board.Comments = res.Comments
	board.ID = res.ID
	board.Ongoing = res.Ongoing
	board.Players = res.Players
	return &board, err
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
