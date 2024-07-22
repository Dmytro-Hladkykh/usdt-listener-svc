package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type CreateUSDTTransferRequest struct {
    FromAddress     string    `json:"from_address"`
    ToAddress       string    `json:"to_address"`
    Amount          string    `json:"amount"`
    TransactionHash string    `json:"transaction_hash"`
    BlockNumber     uint64    `json:"block_number"`
    Timestamp       time.Time `json:"timestamp"`
}

func NewCreateUSDTTransferRequest(r *http.Request) (CreateUSDTTransferRequest, error) {
    var request CreateUSDTTransferRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        return request, errors.Wrap(err, "failed to unmarshal")
    }
    return request, validateCreateUSDTTransferRequest(request)
}

func validateCreateUSDTTransferRequest(request CreateUSDTTransferRequest) error {
    if !common.IsHexAddress(request.FromAddress) {
        return errors.New("invalid 'from' address format")
    }
    if !common.IsHexAddress(request.ToAddress) {
        return errors.New("invalid 'to' address format")
    }
    if request.Amount == "" {
        return errors.New("amount is required")
    }
    if len(request.TransactionHash) != 66 || request.TransactionHash[:2] != "0x" {
        return errors.New("invalid transaction hash format")
    }
    if request.BlockNumber == 0 {
        return errors.New("block number is required")
    }
    if request.Timestamp.IsZero() {
        return errors.New("timestamp is required")
    }
    return nil
}

type ListUSDTTransfersRequest struct {
    Page    int    `json:"page"`
    PerPage int    `json:"per_page"`
    Address string `json:"address"`
}

func NewListUSDTTransfersRequest(r *http.Request) (ListUSDTTransfersRequest, error) {
    request := ListUSDTTransfersRequest{
        Page:    1,
        PerPage: 20,
    }

    query := r.URL.Query()
    if page := query.Get("page"); page != "" {
        if _, err := fmt.Sscanf(page, "%d", &request.Page); err != nil {
            return request, errors.Wrap(err, "invalid page number")
        }
    }
    if perPage := query.Get("per_page"); perPage != "" {
        if _, err := fmt.Sscanf(perPage, "%d", &request.PerPage); err != nil {
            return request, errors.Wrap(err, "invalid per_page number")
        }
    }
    request.Address = query.Get("address")

    return request, validateListUSDTTransfersRequest(request)
}

func validateListUSDTTransfersRequest(request ListUSDTTransfersRequest) error {
    if request.Page < 1 {
        return errors.New("page must be greater than 0")
    }
    if request.PerPage < 1 || request.PerPage > 100 {
        return errors.New("per_page must be between 1 and 100")
    }
    if request.Address != "" && !common.IsHexAddress(request.Address) {
        return errors.New("invalid address format")
    }
    return nil
}