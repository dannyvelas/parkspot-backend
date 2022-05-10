package api

import (
	"context"
	"errors"
)

type userCtxKeyType string

const userCtxKey userCtxKeyType = "user"

type user struct {
	Id   string `json:"id"`
	Role Role   `json:"role"`
}

func ctxWithUser(ctx context.Context, user user) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func ctxGetUser(ctx context.Context) (user, error) {
	if ctxUser := ctx.Value(userCtxKey); ctxUser == nil {
		return user{}, errors.New("user.ctxGetUser: key not found in context")
	} else if parsedUser, ok := ctxUser.(user); !ok {
		return user{}, errors.New("user.ctxGetUser: value for user is not of type `user`")
	} else {
		return parsedUser, nil
	}
}
