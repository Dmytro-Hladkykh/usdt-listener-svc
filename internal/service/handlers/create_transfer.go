package handlers

import (
	"net/http"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func CreateUSDTTransfer(w http.ResponseWriter, r *http.Request) {
	log := Log(r)
	db := DB(r)

	request, err := requests.NewCreateUSDTTransferRequest(r)
	if err != nil {
		log.WithError(err).Error("failed to parse request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	transfer := data.USDTTransfer{
		FromAddress:     request.FromAddress,
		ToAddress:       request.ToAddress,
		Amount:          request.Amount,
		TransactionHash: request.TransactionHash,
		BlockNumber:     request.BlockNumber,
		Timestamp:       request.Timestamp,
	}

	err = db.USDTTransfer().Insert(transfer)
	if err != nil {
		log.WithError(err).Error("failed to create USDT transfer")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, transfer)
}