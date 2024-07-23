package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetUSDTTransfer(w http.ResponseWriter, r *http.Request) {
	log := Log(r)
	db := DB(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.WithError(err).Error("failed to parse id")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	transfer, err := db.USDTTransfer().FilterByID(id).Get()
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
