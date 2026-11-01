package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/skvdmt/jwt"
	erw "github.com/skvdmt/skvdmt-back/pkg/errwrap"
	"github.com/skvdmt/valid-access-back/internal/model"
	"github.com/skvdmt/valid-access-back/internal/repository"
)

// App Сервисный слой.
type App struct {
	// Репозиторный слой.
	repository Repository
}

// NewApp Конструктор.
func NewApp(ctx context.Context) (*App, error) {
	model.Log.Info.Info("usecase layer creating")
	r, err := repository.NewApp(ctx)
	if err != nil {
		return nil, err
	}
	return &App{repository: r}, nil
}

// Stop Остановка.
func (a *App) Stop(ctx context.Context) error {
	model.Log.Info.Info("usecase layer stopped")
	return a.repository.Stop(ctx)
}

// AccessTokenValidation Валидация токена доступа.
func (a *App) AccessTokenValidation(ctx context.Context, token string) error {
	// Валидация токена в кэше.
	if a.repository.AccessTokenValidInCache(ctx, token) {
		return nil
	}
	// Разбор и валидация токена.
	p, err := jwt.Parse(token)
	if err != nil {
		return erw.New(
			erw.CodeHTTP(http.StatusUnauthorized),
		)
	}
	if v, _ := p.Valid(model.Conf.AccessKey); !v {
		return erw.New(
			erw.CodeHTTP(http.StatusUnauthorized),
		)
	}
	// Добавление в кеш.
	e, _ := p.Claim("exp")
	et, _ := time.Parse(time.RFC3339, e)
	u, _ := p.Claim("usr")
	a.repository.AccessTokenAddToCache(ctx, token, u, et.Unix())
	return nil
}
