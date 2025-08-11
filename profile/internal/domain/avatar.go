package domain

import "context"

type Avatar interface {
	AvatarID() string
	Upload(ctx context.Context) error
}
