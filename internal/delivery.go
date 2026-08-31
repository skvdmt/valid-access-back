package internal

import (
	"context"

	valid_access_v1 "github.com/skvdmt/valid-access-back/gen/go/valid_access"
)

// Delivery Интерфейс транспортного слоя.
type Delivery interface {
	// Остановка.
	Stop(context.Context) error
	// Валидация токена доступа.
	AccessTokenValidation(
		context.Context,
		*valid_access_v1.AccessToken,
	) (*valid_access_v1.Empty, error)
}
