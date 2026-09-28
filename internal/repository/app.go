package repository

import (
	"context"
	"sync"
	"time"

	"github.com/skvdmt/valid-access-back/internal/model"
)

const (
	CLEAR_CACHE_INTERVAL = time.Minute * 5
)

// App Резиторный слой.
type App struct {
	accessTokens     *UsersTokenCache
	clearCacheTicker *time.Ticker
	clearCacheExit   chan struct{}
	sources          *sync.WaitGroup
}

// NewApp Конструктор.
func NewApp(_ context.Context) (*App, error) {
	model.Log.Info.Info("repository layer creating")
	a := &App{
		accessTokens:     NewUsersTokenCache(),
		clearCacheTicker: time.NewTicker(CLEAR_CACHE_INTERVAL),
		clearCacheExit:   make(chan struct{}, 1),
		sources:          &sync.WaitGroup{},
	}
	a.sources.Add(1)
	go a.clearCacheHandler()
	return a, nil
}

// Stop Остановка.
func (a *App) Stop(context.Context) error {
	close(a.clearCacheExit)
	a.sources.Wait()
	model.Log.Info.Info("repository layer stopped")
	return nil
}

// AccessTokenValidInCache Валидация токена доступа по кэшу.
func (a *App) AccessTokenValidInCache(_ context.Context, token string) bool {
	return a.accessTokens.Valid(token)
}

// AccessTokenAddToCache Добавление токена доступа в кэш.
func (a *App) AccessTokenAddToCache(_ context.Context, token, username string, exp int64) {
	a.accessTokens.Add(username, token, exp)
}

// clearCacheHandler Обработчик очистки кэша.
func (a *App) clearCacheHandler() {
	model.Log.Info.Info("clear cache handler starting")
	for {
		select {
		case <-a.clearCacheTicker.C:
			a.accessTokens.Clear()
		case <-a.clearCacheExit:
			model.Log.Info.Info("clear cache handler stopped")
			a.sources.Done()
			return
		}
	}
}
