package delivery

import (
	"context"

	valid_access_v1 "github.com/skvdmt/valid-access-back/gen/go/valid_access"
	"github.com/skvdmt/valid-access-back/internal/model"
	"github.com/skvdmt/valid-access-back/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// App Транспортный слой.
type App struct {
	// Сервисный слой.
	usecase Usecase
}

// NewApp Конструктор.
func NewApp(ctx context.Context) (*App, error) {
	model.Log.Info.Info("delivery layer creating")
	a := &App{}
	var err error
	// Создание сервисного слоя.
	a.usecase, err = usecase.NewApp(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Stop Остановка.
func (a *App) Stop(ctx context.Context) error {
	if err := a.usecase.Stop(ctx); err != nil {
		return err
	}
	model.Log.Info.Info("delivery layer stopped")
	return nil
}

// AccessTokenValidation Валидация токена доступа.
func (a *App) AccessTokenValidation(
	ctx context.Context,
	at *valid_access_v1.AccessToken,
) (*valid_access_v1.Empty, error) {
	err := a.usecase.AccessTokenValidation(ctx, at.Token)
	if err != nil {
		return &valid_access_v1.Empty{}, status.Error(
			codes.Unauthenticated, "invalid access token")
	}
	return &valid_access_v1.Empty{}, nil
}
