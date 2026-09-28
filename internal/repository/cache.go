package repository

import (
	"sync"
	"time"
)

// TokenInfo Информация о токене.
type TokenInfo struct {
	// Токен.
	Token string
	// Срок действия истекает.
	Expire int64
}

// UsersTokenCache Кэш токенов пользователей.
type UsersTokenCache struct {
	storage map[string][]*TokenInfo
	mu      *sync.RWMutex
}

// NewUsersTokenCache Конструктор.
func NewUsersTokenCache() *UsersTokenCache {
	return &UsersTokenCache{
		storage: make(map[string][]*TokenInfo),
		mu:      &sync.RWMutex{},
	}
}

// Add Добавление токена.
func (u *UsersTokenCache) Add(username, token string, exp int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	// Если закэшировано более 9 токенов пользователя, удалить их все.
	if len(u.storage[username]) > 9 {
		u.storage[username] = []*TokenInfo{}
	}
	u.storage[username] = append(u.storage[username], &TokenInfo{
		Token:  token,
		Expire: exp,
	})
}

// Valid Токен валиден.
func (u *UsersTokenCache) Valid(token string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	for _, ts := range u.storage {
		for _, t := range ts {
			if t.Token == token {
				if t.Expire < time.Now().Unix() {
					return false
				}
				return true
			}
		}
	}
	return false
}

// Clear Очистка.
func (u *UsersTokenCache) Clear() {
	n := make(map[string][]*TokenInfo)
	u.mu.Lock()
	defer u.mu.Unlock()
	for u, ts := range u.storage {
		var i []*TokenInfo
		for _, t := range ts {
			if t.Expire > time.Now().Unix() {
				i = append(i, t)
			}
		}
		if len(i) > 0 {
			n[u] = i
		}
	}
	u.storage = n
}
