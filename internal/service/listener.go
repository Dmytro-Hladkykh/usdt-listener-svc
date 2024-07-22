// File: internal/service/listener.go

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
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const (
    USDTContractAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7" // USDT contract address on Ethereum mainnet
    TransferEventSignature = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" // keccak256("Transfer(address,address,uint256)")
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
        Topics:    [][]common.Hash{{common.HexToHash(TransferEventSignature)}},
    }

    logs := make(chan types.Log)
    sub, err := l.client.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return errors.Wrap(err, "failed to subscribe to logs")
    }

    for {
        select {
        case err := <-sub.Err():
            return errors.Wrap(err, "subscription error")
        case vLog := <-logs:
            if err := l.processLog(vLog); err != nil {
                l.log.WithError(err).Error("failed to process log")
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (l *Listener) processLog(vLog types.Log) error {
    if len(vLog.Topics) != 3 {
        return errors.New("invalid number of topics in log")
    }

    fromAddress := common.HexToAddress(vLog.Topics[1].Hex())
    toAddress := common.HexToAddress(vLog.Topics[2].Hex())
    amount := new(big.Int).SetBytes(vLog.Data)

    transfer := data.USDTTransfer{
        FromAddress:     fromAddress.Hex(),
        ToAddress:       toAddress.Hex(),
        Amount:          amount.String(),
        TransactionHash: vLog.TxHash.Hex(),
        BlockNumber:     vLog.BlockNumber,
        Timestamp:       time.Now().UTC(),
    }

    _, err := l.db.USDTTransfer().Insert(transfer)
    if err != nil {
        return errors.Wrap(err, "failed to insert USDT transfer")
    }

    l.log.WithFields(logan.F{
        "from":   transfer.FromAddress,
        "to":     transfer.ToAddress,
        "amount": transfer.Amount,
        "txHash": transfer.TransactionHash,
    }).Info("USDT transfer processed")

    return nil
}