package handlers

import (
	"net/http"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/service/requests"
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

	transfersQ := db.USDTTransfer()

	if request.Address != "" {
		transfersQ = transfersQ.FilterByFromAddress(request.Address)
	}

	page := uint64(request.Page)
	perPage := uint64(request.PerPage)
	offset := (page - 1) * perPage

	transfersQ = transfersQ.Limit(perPage).Offset(offset)

	transfers, err := transfersQ.Select()
	if err != nil {
		log.WithError(err).Error("failed to get USDT transfers")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, transfers)
}
