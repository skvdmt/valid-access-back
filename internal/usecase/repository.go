package usecase

import (
	"context"
)

// Repository Интерфейс репозиторного слоя.
type Repository interface {
	// Остановка.
	Stop(context.Context) error
	// Валидация токена доступа по кэшу.
	AccessTokenValidInCache(c context.Context, token string) bool
	// Добавление токена доступа в кэш.
	AccessTokenAddToCache(c context.Context, token, username string, exp int64)
}
