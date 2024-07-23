package service

import (
	"context"
	"math/big"
	"time"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const (
    USDTContractAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
)

type Listener struct {
    client *ethclient.Client
    db     data.MasterQ
    log    *logan.Entry
}

func NewListener(infuraURL string, db data.MasterQ, log *logan.Entry) (*Listener, error) {
    client, err := ethclient.Dial(infuraURL)
    if err != nil {
        return nil, errors.Wrap(err, "failed to connect to Ethereum client")
    }

    return &Listener{
        client: client,
        db:     db,
        log:    log,
    }, nil
}

func (l *Listener) Listen(ctx context.Context) error {
    contractAddress := common.HexToAddress(USDTContractAddress)
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
    }

    logs := make(chan types.Log)
    sub, err := l.client.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return errors.Wrap(err, "failed to subscribe to logs")
    }

    l.log.Info("Started listening for USDT transfers")

    for {
        select {
        case err := <-sub.Err():
            return errors.Wrap(err, "subscription error")
        case vLog := <-logs:
            if err := l.processLog(vLog); err != nil {
                l.log.WithError(err).Error("Error processing log")
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (l *Listener) processLog(vLog types.Log) error {
    if len(vLog.Topics) != 3 {
        return nil 
    }

    from := common.HexToAddress(vLog.Topics[1].Hex())
    to := common.HexToAddress(vLog.Topics[2].Hex())
    amount := new(big.Int).SetBytes(vLog.Data)

    decimalAmount := decimal.NewFromBigInt(amount, 0).Div(decimal.New(1, 6))

    block, err := l.client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
    if err != nil {
        return errors.Wrap(err, "failed to get block information")
    }

    transfer := data.USDTTransfer{
        FromAddress:     from.Hex(),
        ToAddress:       to.Hex(),
        Amount:          decimalAmount.String(),
        TransactionHash: vLog.TxHash.Hex(),
        BlockNumber:     vLog.BlockNumber,
        Timestamp:       time.Unix(int64(block.Time()), 0), 
    }

    _, err = l.db.USDTTransfer().Insert(transfer)
    if err != nil {
        return errors.Wrap(err, "failed to insert USDT transfer")
    }

    l.log.WithFields(logan.F{
        "from":      transfer.FromAddress,
        "to":        transfer.ToAddress,
        "amount":    transfer.Amount,
        "txHash":    transfer.TransactionHash,
        "timestamp": transfer.Timestamp,
    }).Info("New USDT transfer processed")

    return nil
}