package api

import (
	"context"
	"errors"
)

type userCtxKeyType string

const userCtxKey userCtxKeyType = "user"

type user struct {
	Id        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Role      Role   `json:"role"`
}

func newUser(id string, firstName string, lastName string, email string, role Role) user {
	return user{
		Id:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Role:      role,
	}
}

func (self user) Equal(other user) bool {
	if self.Id != other.Id {
		return false
	} else if self.FirstName != other.FirstName {
		return false
	} else if self.LastName != other.LastName {
		return false
	} else if self.Email != other.Email {
		return false
	} else if self.Role != other.Role {
		return false
	}
	return true
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
