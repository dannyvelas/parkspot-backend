package api

import (
	"context"

	"github.com/dannyvelas/parkspot-backend/app"
)

type accessKeyType string

const accessKey accessKeyType = "access"

func ctxWithAccessPayload(ctx context.Context, payload app.AccessPayload) context.Context {
	return context.WithValue(ctx, accessKey, payload)
}

func ctxGetAccessPayload(ctx context.Context) (app.AccessPayload, error) {
	if ctxUser := ctx.Value(accessKey); ctxUser == nil {
		return app.AccessPayload{}, ErrAccessPayloadNotInContext
	} else if payload, ok := ctxUser.(app.AccessPayload); !ok {
		return app.AccessPayload{}, ErrAccessPayloadWrongType
	} else {
		return payload, nil
	}
}
