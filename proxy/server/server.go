package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/visonlv/go-vkit/logger"
)

type Server interface {
	Start() error
	Stop() error
}

type Config struct {
	Host         string `env:"HOST"            envDefault:""`
	Port         string `env:"PORT"            envDefault:""`
	CertFile     string `env:"SERVER_CERT"     envDefault:""`
	KeyFile      string `env:"SERVER_KEY"      envDefault:""`
	ServerCAFile string `env:"SERVER_CA_CERTS" envDefault:""`
	ClientCAFile string `env:"CLIENT_CA_CERTS" envDefault:""`
}

type BaseServer struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Name     string
	Address  string
	Config   *Config
	Protocol string
}

func stopAllServer(servers ...Server) error {
	var err error
	for _, server := range servers {
		err1 := server.Stop()
		if err1 != nil {
			if err == nil {
				err = fmt.Errorf("%w", err1)
			} else {
				err = fmt.Errorf("%v ; %w", err, err1)
			}
		}
	}
	return err
}

func StopSignalHandler(ctx context.Context, cancel context.CancelFunc, svcName string, servers ...Server) error {
	var err error
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	select {
	case sig := <-c:
		defer cancel()
		err = stopAllServer(servers...)
		if err != nil {
			logger.Errorf("%s service error during shutdown: %v", svcName, err)
		}
		logger.Infof("%s service shutdown by signal: %s", svcName, sig)
		return err
	case <-ctx.Done():
		return nil
	}
}
