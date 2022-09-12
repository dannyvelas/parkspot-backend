package api

import (
	"context"
	"errors"
)

type accessKeyType string

const accessKey accessKeyType = "access"

func ctxWithAccessPayload(ctx context.Context, payload accessPayload) context.Context {
	return context.WithValue(ctx, accessKey, payload)
}

func ctxGetAccessPayload(ctx context.Context) (accessPayload, error) {
	if ctxUser := ctx.Value(accessKey); ctxUser == nil {
		return accessPayload{}, errors.New("user.ctxGetUser: key not found in context")
	} else if payload, ok := ctxUser.(accessPayload); !ok {
		return accessPayload{}, errors.New("user.ctxGetUser: value for user is not of type `user`")
	} else {
		return payload, nil
	}
}
