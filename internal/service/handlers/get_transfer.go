package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetUSDTTransfer(w http.ResponseWriter, r *http.Request) {
	log := Log(r)
	db := DB(r)

	id := chi.URLParam(r, "id")

	transfer, err := db.USDTTransfer().FilterById(id).Get()
	if err != nil {
		log.WithError(err).Error("failed to get USDT transfer")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if transfer == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, transfer)
}