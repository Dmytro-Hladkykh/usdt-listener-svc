package requests

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/urlval"
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
    Page    int    `url:"page"`
    PerPage int    `url:"per_page"`
    Address string `url:"address"`
    Limit   uint64
    PageNumber uint64
}

func NewListUSDTTransfersRequest(r *http.Request) (ListUSDTTransfersRequest, error) {
    var request ListUSDTTransfersRequest

    err := urlval.Decode(r.URL.Query(), &request)
    if err != nil {
        return request, errors.Wrap(err, "failed to decode query parameters")
    }

    if request.Page == 0 {
        request.Page = 1
    }
    if request.PerPage == 0 {
        request.PerPage = 20
    }

    request.Limit = uint64(request.PerPage)
    request.PageNumber = uint64(request.Page)

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

func (r ListUSDTTransfersRequest) GetPageParams() pgdb.OffsetPageParams {
    return pgdb.OffsetPageParams{
        Limit:      r.Limit,
        PageNumber: r.PageNumber,
    }
}
