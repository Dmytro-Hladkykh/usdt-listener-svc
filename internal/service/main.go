package service

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/config"
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data/pg"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type service struct {
    log      *logan.Entry
    copus    types.Copus
    listener net.Listener
    cfg      config.Config
}

func (s *service) run(cfg config.Config) error {
    s.log.Info("Service started")
    r := s.router(cfg)

    if err := s.copus.RegisterChi(r); err != nil {
        return errors.Wrap(err, "cop failed")
    }

    // Start the USDT listener
    go s.runUSDTListener()

    return http.Serve(s.listener, r)
}

func (s *service) runUSDTListener() {
    infuraAPIKey := os.Getenv("INFURA_API_KEY")
    if infuraAPIKey == "" {
        s.log.Error("INFURA_API_KEY environment variable is not set")
        return
    }
    
    infuraURL := "wss://mainnet.infura.io/ws/v3/" + infuraAPIKey
    db := pg.NewMasterQ(s.cfg.DB())
    
    usdtListener, err := NewListener(infuraURL, db, s.log)
    if err != nil {
        s.log.WithError(err).Error("Failed to create USDT listener")
        return
    }

    if err := usdtListener.Listen(context.Background()); err != nil {
        s.log.WithError(err).Error("USDT listener stopped")
    }
}

func newService(cfg config.Config) *service {
    return &service{
        log:      cfg.Log(),
        copus:    cfg.Copus(),
        listener: cfg.Listener(),
        cfg:      cfg,
    }
}

func Run(cfg config.Config) {
    if err := newService(cfg).run(cfg); err != nil {
        panic(err)
    }
}