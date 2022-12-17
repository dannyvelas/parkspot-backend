package api

import (
	"context"
	"errors"
	"github.com/dannyvelas/lasvistas_api/app"
)

type accessKeyType string

const accessKey accessKeyType = "access"

func ctxWithAccessPayload(ctx context.Context, payload app.AccessPayload) context.Context {
	return context.WithValue(ctx, accessKey, payload)
}

func ctxGetAccessPayload(ctx context.Context) (app.AccessPayload, error) {
	if ctxUser := ctx.Value(accessKey); ctxUser == nil {
		return app.AccessPayload{}, errors.New("user.ctxGetUser: key not found in context")
	} else if payload, ok := ctxUser.(app.AccessPayload); !ok {
		return app.AccessPayload{}, errors.New("user.ctxGetUser: value for user is not of type `user`")
	} else {
		return payload, nil
	}
}
