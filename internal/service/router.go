package service

import (
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/config"
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data/pg"
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/service/handlers"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (s *service) router(cfg config.Config) chi.Router {
  r := chi.NewRouter()

  r.Use(
    ape.RecoverMiddleware(s.log),
    ape.LoganMiddleware(s.log),
    ape.CtxMiddleware(
      handlers.CtxLog(s.log),
      handlers.CtxDB(pg.NewMasterQ(cfg.DB())),
    ),
  )
  r.Route("/usdt-listener-svc", func(r chi.Router) {
      r.Get("/", handlers.ListUSDTTransfers)
      r.Get("/{id}", handlers.GetUSDTTransfer)
  })

  return r
}
