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
	accessTokens     map[string]int64
	mu               *sync.RWMutex
	clearCacheTicker *time.Ticker
	clearCacheExit   chan struct{}
	sources          *sync.WaitGroup
}

// NewApp Конструктор.
func NewApp(ctx context.Context) (*App, error) {
	model.Log.Info.Info("repository layer creating")
	a := &App{
		accessTokens:     make(map[string]int64),
		mu:               &sync.RWMutex{},
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
func (a *App) AccessTokenValidInCache(c context.Context, token string) bool {
	a.mu.RLock()
	t, ok := a.accessTokens[token]
	a.mu.RUnlock()
	if !ok {
		return false
	}
	if t < time.Now().Unix() {
		return false
	}
	return true
}

// AccessTokenAddToCache Добавление токена доступа в кэш.
func (a *App) AccessTokenAddToCache(c context.Context, token string, exp int64) {
	a.mu.Lock()
	a.accessTokens[token] = exp
	a.mu.Unlock()
}

// clearCacheHandler Обработчик очистки кэша.
func (a *App) clearCacheHandler() {
	model.Log.Info.Info("clear cache handler starting")
	for {
		select {
		case <-a.clearCacheTicker.C:
			a.clearCache()
		case <-a.clearCacheExit:
			model.Log.Info.Info("clear cache handler stopped")
			a.sources.Done()
			return
		}
	}
}

// clearCache Очистка кеша.
func (a *App) clearCache() {
	model.Log.Info.Info("clear cache")
	n := make(map[string]int64)
	a.mu.Lock()
	defer a.mu.Unlock()
	for t, e := range a.accessTokens {
		if e > time.Now().Unix() {
			n[t] = e
		}
	}
	a.accessTokens = n
}
