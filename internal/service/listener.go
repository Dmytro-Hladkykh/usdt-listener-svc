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
    USDTContractAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
    ReconnectDelay      = 5 * time.Second
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
    for {
        if err := l.processHistoricalEvents(ctx); err != nil {
            l.log.WithError(err).Error("Failed to process historical events")
            time.Sleep(ReconnectDelay)
            continue
        }

        if err := l.listenForNewEvents(ctx); err != nil {
            l.log.WithError(err).Error("Error listening for new events")
            time.Sleep(ReconnectDelay)
            continue
        }

        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue the loop
        }
    }
}

func (l *Listener) processHistoricalEvents(ctx context.Context) error {
    lastProcessedBlock, err := l.db.LastProcessedBlock().Get()
    if err != nil {
        return errors.Wrap(err, "failed to get last processed block")
    }

    currentBlock, err := l.client.BlockNumber(ctx)
    if err != nil {
        return errors.Wrap(err, "failed to get current block number")
    }

    if lastProcessedBlock < currentBlock {
        l.log.WithFields(logan.F{
            "lastProcessedBlock": lastProcessedBlock,
            "currentBlock":       currentBlock,
        }).Info("Processing missed blocks")

        contractAddress := common.HexToAddress(USDTContractAddress)

        for blockNum := lastProcessedBlock + 1; blockNum <= currentBlock; blockNum++ {
            l.log.WithFields(logan.F{
                "blockNumber": blockNum,
            }).Info("Processing block")

            query := ethereum.FilterQuery{
                FromBlock: big.NewInt(int64(blockNum)),
                ToBlock:   big.NewInt(int64(blockNum)),
                Addresses: []common.Address{contractAddress},
            }

            logs, err := l.client.FilterLogs(ctx, query)
            if err != nil {
                l.log.WithFields(logan.F{
                    "blockNumber": blockNum,
                }).WithError(err).Error("Failed to filter logs")
                continue
            }

            transactions := make([]data.USDTTransfer, 0)
            for _, vLog := range logs {
                tx, err := l.extractTransaction(vLog)
                if err != nil {
                    l.log.WithFields(logan.F{
                        "blockNumber": blockNum,
                        "logIndex":    vLog.Index,
                    }).WithError(err).Error("Error processing historical log")
                    continue
                }
                transactions = append(transactions, tx)
            }

            if len(transactions) > 0 {
                if err := l.db.USDTTransfer().InsertBlock(transactions); err != nil {
                    l.log.WithFields(logan.F{
                        "blockNumber": blockNum,
                    }).WithError(err).Error("Failed to insert block of USDT transfers")
                    continue
                }
            }

            if err := l.db.LastProcessedBlock().Update(blockNum); err != nil {
                l.log.WithFields(logan.F{
                    "blockNumber": blockNum,
                }).WithError(err).Error("Failed to update last processed block")
                continue
            }

            l.log.WithFields(logan.F{
                "blockNumber": blockNum,
            }).Info("Successfully processed block")
        }
    }

    return nil
}

func (l *Listener) listenForNewEvents(ctx context.Context) error {
    contractAddress := common.HexToAddress(USDTContractAddress)
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
    }
    logs := make(chan types.Log)
    sub, err := l.client.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return errors.Wrap(err, "failed to subscribe to logs")
    }
    l.log.Info("Started listening for new USDT transfers")
    for {
        select {
        case err := <-sub.Err():
            return errors.Wrap(err, "subscription error")
        case vLog := <-logs:
            tx, err := l.extractTransaction(vLog)
            if err != nil {
                l.log.WithError(err).Error("Error processing log")
                continue
            }
            if _, err := l.db.USDTTransfer().Insert(tx); err != nil {
                l.log.WithError(err).Error("Failed to insert transaction")
            } else {
                l.log.WithFields(logan.F{
                    "from":      tx.FromAddress,
                    "to":        tx.ToAddress,
                    "amount":    tx.Amount,
                    "txHash":    tx.TransactionHash,
                    "timestamp": tx.Timestamp,
                }).Info("New USDT transfer processed")
            }
            if err := l.db.LastProcessedBlock().Update(vLog.BlockNumber); err != nil {
                l.log.WithError(err).Error("Failed to update last processed block")
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (l *Listener) extractTransaction(vLog types.Log) (data.USDTTransfer, error) {
    if len(vLog.Topics) != 3 {
        return data.USDTTransfer{}, errors.New("invalid log topics length")
    }

    from := common.HexToAddress(vLog.Topics[1].Hex())
    to := common.HexToAddress(vLog.Topics[2].Hex())
    amount := new(big.Int).SetBytes(vLog.Data)

    block, err := l.client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
    if err != nil {
        return data.USDTTransfer{}, errors.Wrap(err, "failed to get block information")
    }

    return data.USDTTransfer{
        FromAddress:     from.Hex(),
        ToAddress:       to.Hex(),
        Amount:          amount.String(),
        TransactionHash: vLog.TxHash.Hex(),
        BlockNumber:     vLog.BlockNumber,
        LogIndex:        uint64(vLog.Index),
        Timestamp:       time.Unix(int64(block.Time()), 0),
    }, nil
}
