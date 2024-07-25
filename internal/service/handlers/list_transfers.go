package handlers

import (
	"net/http"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/service/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/kit/pgdb"
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

    pageParams := pgdb.OffsetPageParams{
        Limit:  uint64(request.PerPage),
		PageNumber: uint64(request.Page),
    }

    transfers, err := transfersQ.Page(&pageParams).Select()
    if err != nil {
        log.WithError(err).Error("failed to get USDT transfers")
        ape.RenderErr(w, problems.InternalError())
        return
    }

    ape.Render(w, transfers)
}