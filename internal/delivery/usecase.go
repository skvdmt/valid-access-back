package delivery

import (
	"context"
)

// Usecase Интерфейс сервисного слоя.
type Usecase interface {
	// Остановка.
	Stop(context.Context) error
	// Валидация токена доступа.
	AccessTokenValidation(c context.Context, token string) error
}
