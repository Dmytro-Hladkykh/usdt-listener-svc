package handlers

import (
	"net/http"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func ListUSDTTransfers(w http.ResponseWriter, r *http.Request) {
	log := Log(r)
	db := DB(r)

	request, err := requests.NewListUSDTTransfersRequest(r)
	if err != nil {
		log.WithError(err).Error("failed to parse request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	transfers, err := db.USDTTransfer().
		Page(request.PageParams).
		FilterByAddress(request.Address).
		Select()

	if err != nil {
		log.WithError(err).Error("failed to get USDT transfers")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, transfers)
}