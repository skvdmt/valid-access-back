package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	valid_access_v1 "github.com/skvdmt/valid-access-back/gen/go/valid_access"
	"github.com/skvdmt/valid-access-back/internal/delivery"
	"github.com/skvdmt/valid-access-back/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	CERT_DIRECTORY_DEV  = "./certs/valid-access-back"
	CERT_DIRECTORY_PROD = "/etc/ssl/certs/valid-access-back"
)

// App Основная структура приложения.
type App struct {
	// Канал сигналов остановки.
	interrupt chan os.Signal
	// Контекст приложения.
	ctx context.Context
	// Отмена контекста.
	cancel context.CancelFunc
	// Транспортный слой.
	delivery Delivery
	// Сервер валидации.
	server *grpc.Server
	// Ошибки приложения.
	errors []error
	// Приложение уже закрывается.
	stopping bool
	// Корректное завершение горутин.
	wg *sync.WaitGroup
}

// NewApp Конструктор.
func NewApp() (*App, error) {
	model.Log.Info.Info(fmt.Sprintf("app %s creating", model.NAME))
	a := &App{
		interrupt: make(chan os.Signal),
		wg:        &sync.WaitGroup{},
	}
	var err error
	// Загрузка конфигурации.
	if model.Conf, err = model.NewConfig(); err != nil {
		return nil, err
	}
	// Создане глобального канала ошибок для всего приложения.
	model.Errors = make(chan error)
	// Создание контекста.
	a.ctx, a.cancel = context.WithCancel(context.Background())
	// Создание транспортного слоя.
	a.delivery, err = delivery.NewApp(a.ctx)
	if err != nil {
		return nil, err
	}
	// Создание сервера валидации токена доступа.
	c, err := credentials.NewServerTLSFromFile(
		filepath.Join(a.certsDir(), "cert.crt"),
		filepath.Join(a.certsDir(), "private.key"),
	)
	if err != nil {
		return nil, err
	}
	a.server = grpc.NewServer(grpc.Creds(c))
	valid_access_v1.RegisterValidatorServer(a.server, a.delivery)
	return a, nil
}

// certsDir Директория сертификатов.
func (a *App) certsDir() string {
	m, o := os.LookupEnv(model.MODE)
	if o && m == model.MODE_DEV {
		return CERT_DIRECTORY_DEV
	}
	return CERT_DIRECTORY_PROD
}

// Start Запуск приложения.
func (a *App) Start() []error {
	model.Log.Info.Info(fmt.Sprintf("app %s starting", model.NAME))

	// Запуск обработчика ошибок.
	a.wg.Add(1)
	go a.errorHandler()

	// Запуск сервера валидации.
	a.wg.Add(1)
	go a.validServerStart()

	return a.interruptHandler()
}

// errorHandler Обработчик глобального канала ошибок.
func (a *App) errorHandler() {
	model.Log.Info.Info("error handle starting")
	defer func() {
		model.Log.Info.Info("error handle stopped")
		a.wg.Done()
	}()
	for {
		err := <-model.Errors
		if err == nil {
			return
		}
		if errors.Is(err, context.Canceled) {
			continue
		}
		a.errors = append(a.errors, err)
		if !a.stopping {
			a.interrupt <- syscall.SIGTERM
		}
	}
}

// validServerStart Запуск сервера валидации.
func (a *App) validServerStart() {
	model.Log.Info.Info(fmt.Sprintf("validation server starting on %d port", model.Conf.Server.Port))
	defer func() {
		model.Log.Info.Info("validation gRPC server stopped")
		a.wg.Done()
	}()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", model.Conf.Server.Port))
	if err != nil {
		model.Errors <- err
		return
	}
	err = a.server.Serve(lis)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		model.Errors <- err
	}
}

// interruptHandler Обработчик сигналов остановки приложения.
func (a *App) interruptHandler() []error {
	model.Log.Info.Info("waiting for interrupt signals")
	signal.Notify(a.interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-a.interrupt
	a.stop()
	// Закрытие канала ошибок.
	close(model.Errors)
	// Ожидание завершения горутин.
	a.wg.Wait()
	return a.errors
}

// stop Остановка.
func (a *App) stop() {
	defer model.Log.Info.Info(fmt.Sprintf("app %s stopped", model.NAME))
	a.stopping = true
	// Отмена контекста.
	a.cancel()
	model.Log.Info.Info("context canceled")
	// Остановка сервера валидации.
	a.server.GracefulStop()
	// Остановка транспортного слоя.
	if err := a.delivery.Stop(a.ctx); err != nil {
		model.Errors <- err
	}
	// Закрытие канала сигналов прерывания.
	close(a.interrupt)
}
