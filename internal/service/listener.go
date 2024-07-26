package service

import (
	"context"
	"math/big"
	"os"
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
    ReconnectDelay      = 15 * time.Second
    MaxRetryAttempts    = 5
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
    processHist := os.Getenv("PROCESS_HIST")

    var lastProcessedBlock uint64
    var err error

    if processHist == "true" {
        lastProcessedBlock, err = l.processHistoricalEvents(ctx)
        if err != nil {
            l.log.WithError(err).Error("Failed to process historical events")
        }
    } else {
        l.log.Info("Skipping historical events processing")
        currentBlock, err := l.client.BlockNumber(ctx)
        if err != nil {
            return errors.Wrap(err, "failed to get current block number")
        }
        lastProcessedBlock = currentBlock
        if err := l.db.LastProcessedBlock().Update(currentBlock); err != nil {
            l.log.WithError(err).Error("Failed to update last processed block")
        }
    }

    return l.listenForNewEvents(ctx, lastProcessedBlock)
}

func (l *Listener) processHistoricalEvents(ctx context.Context) (uint64, error) {
    lastProcessedBlock, err := l.db.LastProcessedBlock().Get()
    if err != nil {
        return 0, errors.Wrap(err, "failed to get last processed block")
    }

    for {
        currentBlock, err := l.client.BlockNumber(ctx)
        if err != nil {
            return lastProcessedBlock, errors.Wrap(err, "failed to get current block number")
        }

        if lastProcessedBlock >= currentBlock {
            break
        }

        l.log.WithFields(logan.F{
            "lastProcessedBlock": lastProcessedBlock,
            "currentBlock":       currentBlock,
        }).Info("Processing historical blocks")

        batchSize := uint64(10)
        for blockNum := lastProcessedBlock + 1; blockNum <= currentBlock; blockNum += batchSize {
            endBlock := blockNum + batchSize - 1
            if endBlock > currentBlock {
                endBlock = currentBlock
            }

            l.log.WithFields(logan.F{
                "fromBlock": blockNum,
                "toBlock":   endBlock,
            }).Info("Processing block range")

            if err := l.processBlockRange(ctx, blockNum, endBlock); err != nil {
                l.log.WithError(err).Error("Failed to process block range")
                time.Sleep(ReconnectDelay)
                continue
            }

            lastProcessedBlock = endBlock
            if err := l.db.LastProcessedBlock().Update(lastProcessedBlock); err != nil {
                l.log.WithError(err).Error("Failed to update last processed block")
            }
        }
    }

    return lastProcessedBlock, nil
}

func (l *Listener) processBlockRange(ctx context.Context, startBlock, endBlock uint64) error {
    for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
        if err := l.processBlock(ctx, blockNum); err != nil {
            l.log.WithError(err).WithField("blockNumber", blockNum).Error("Failed to process block")
            continue
        }

        if err := l.db.Transaction(func(q data.MasterQ) error {
            if err := q.LastProcessedBlock().Update(blockNum); err != nil {
                return errors.Wrap(err, "failed to update last processed block")
            }
            return nil
        }); err != nil {
            l.log.WithError(err).WithField("blockNumber", blockNum).Error("Failed to update last processed block")
        }
    }
    return nil
}

func (l *Listener) processBlock(ctx context.Context, blockNum uint64) error {
    contractAddress := common.HexToAddress(USDTContractAddress)
    query := ethereum.FilterQuery{
        FromBlock: big.NewInt(int64(blockNum)),
        ToBlock:   big.NewInt(int64(blockNum)),
        Addresses: []common.Address{contractAddress},
    }

    logs, err := l.client.FilterLogs(ctx, query)
    if err != nil {
        return errors.Wrap(err, "failed to filter logs")
    }

    transactions := make([]data.USDTTransfer, 0, len(logs))
    for _, vLog := range logs {
        tx, err := l.extractTransaction(vLog)
        if err != nil {
            l.log.WithFields(logan.F{
                "blockNumber": vLog.BlockNumber,
                "logIndex":    vLog.Index,
            }).WithError(err).Error("Error processing log")
            continue
        }
        transactions = append(transactions, tx)
    }

    if len(transactions) > 0 {
        if err := l.db.USDTTransfer().InsertBlock(transactions); err != nil {
            return errors.Wrap(err, "failed to insert block of USDT transfers")
        }
    }

    return nil
}

func (l *Listener) listenForNewEvents(ctx context.Context, startBlock uint64) error {
    for {
        currentBlock, err := l.client.BlockNumber(ctx)
        if err != nil {
            l.log.WithError(err).Error("Failed to get current block number")
            time.Sleep(ReconnectDelay)
            continue
        }

        if startBlock < currentBlock {
            if err := l.processBlockRange(ctx, startBlock+1, currentBlock); err != nil {
                l.log.WithError(err).Error("Failed to process missed blocks")
                time.Sleep(ReconnectDelay)
                continue
            }
            startBlock = currentBlock
        }

        if err := l.subscribeToNewLogs(ctx, startBlock+1); err != nil {
            l.log.WithError(err).Error("Error in log subscription")
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

func (l *Listener) subscribeToNewLogs(ctx context.Context, fromBlock uint64) error {
    contractAddress := common.HexToAddress(USDTContractAddress)
    query := ethereum.FilterQuery{
        FromBlock: big.NewInt(int64(fromBlock)),
        Addresses: []common.Address{contractAddress},
    }
    logs := make(chan types.Log)
    sub, err := l.client.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return errors.Wrap(err, "failed to subscribe to logs")
    }
    defer sub.Unsubscribe()

    l.log.WithField("fromBlock", fromBlock).Info("Started listening for new USDT transfers")

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
