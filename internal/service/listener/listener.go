package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/config"
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
    BlockTime           = 12 * time.Second // Ethereum block time
)

// Listener struct
type Listener struct {
    client *ethclient.Client
    db     data.MasterQ
    log    *logan.Entry
    config config.Config
}

// NewListener creates a new Listener instance
func NewListener(config config.Config, db data.MasterQ, log *logan.Entry) (*Listener, error) {
    client, err := ethclient.Dial(config.Ethereum().RPCURL)
    if err != nil {
        return nil, errors.Wrap(err, "failed to connect to Ethereum client")
    }
    return &Listener{
        client: client,
        db:     db,
        log:    log,
        config: config,
    }, nil
}

// Listen starts the main loop for listening to USDT transfers
func (l *Listener) Listen(ctx context.Context, processHist bool, configStartingBlock uint64) error {
    startBlock, err := l.getStartingBlock(ctx, configStartingBlock)
    if err != nil {
        return errors.Wrap(err, "failed to get starting block")
    }

    l.log.WithFields(logan.F{
        "configStartingBlock": configStartingBlock,
        "actualStartingBlock": startBlock,
    }).Info("Starting USDT listener")

    return l.processBlocks(ctx, startBlock)
}

// getStartingBlock determines the block to start processing from
func (l *Listener) getStartingBlock(ctx context.Context, configStartingBlock uint64) (uint64, error) {
    dbBlock, err := l.db.LastProcessedBlock().Get()
    if err != nil {
        return 0, errors.Wrap(err, "failed to get last processed block from DB")
    }

    startingBlock := max(configStartingBlock, dbBlock)

    currentBlock, err := l.client.BlockNumber(ctx)
    if err != nil {
        return 0, errors.Wrap(err, "failed to get current block number")
    }

    return min(startingBlock, currentBlock), nil
}

// processBlocks continuously processes blocks
func (l *Listener) processBlocks(ctx context.Context, startBlock uint64) error {
    l.log.WithField("startingBlock", startBlock).Info("Starting to process blocks")

    for {
        currentBlock, err := l.client.BlockNumber(ctx)
        if err != nil {
            l.log.WithError(err).Error("Failed to get current block number")
            time.Sleep(BlockTime)
            continue
        }

        // Double-check that we're starting from the correct block
        lastProcessedBlock, err := l.db.LastProcessedBlock().Get()
        if err != nil {
            l.log.WithError(err).Error("Failed to get last processed block from DB")
            time.Sleep(time.Second)
            continue
        }

        if lastProcessedBlock >= startBlock {
            l.log.WithFields(logan.F{
                "lastProcessedBlock": lastProcessedBlock,
                "startBlock":         startBlock,
            }).Info("Resuming from last processed block")
            startBlock = lastProcessedBlock + 1
        }

        // Log processing only if we're actually processing a block
        l.log.WithFields(logan.F{
            "currentNetworkBlock": currentBlock,
            "processingBlock":     startBlock,
        }).Info("Processing block")

        if err := l.processBlock(ctx, startBlock); err != nil {
            l.log.WithError(err).WithField("blockNumber", startBlock).Error("Failed to process block")
            time.Sleep(time.Second)
            continue
        }

        // Increment the block number after successful processing
        startBlock++

        // Update the last processed block
        if err := l.db.LastProcessedBlock().Update(startBlock); err != nil {
            l.log.WithError(err).Error("Failed to update last processed block")
        }

        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue processing
        }
    }
}

// processBlock processes a single block
func (l *Listener) processBlock(ctx context.Context, blockNum uint64) error {
    return l.db.Transaction(func(q data.MasterQ) error {
        // Get logs for the block
        logs, err := l.getBlockLogs(ctx, blockNum)
        if err != nil {
            return errors.Wrap(err, "failed to get block logs")
        }

        // Convert logs to USDT transfers
        transfers, err := l.logsToTransfers(ctx, logs, blockNum)
        if err != nil {
            return errors.Wrap(err, "failed to convert logs to transfers")
        }

        // Insert transfers into the database
        for _, transfer := range transfers {
            _, err := q.USDTTransfer().Insert(transfer)
            if err != nil {
                return errors.Wrap(err, "failed to insert transfer")
            }
        }

        // Update the last processed block
        if err := q.LastProcessedBlock().Update(blockNum); err != nil {
            return errors.Wrap(err, "failed to update last processed block")
        }

        return nil
    })
}

// getBlockLogs retrieves logs for a specific block
func (l *Listener) getBlockLogs(ctx context.Context, blockNum uint64) ([]types.Log, error) {
    contractAddress := common.HexToAddress(USDTContractAddress)
    query := ethereum.FilterQuery{
        FromBlock: big.NewInt(int64(blockNum)),
        ToBlock:   big.NewInt(int64(blockNum)),
        Addresses: []common.Address{contractAddress},
    }

    logs, err := l.client.FilterLogs(ctx, query)
    if err != nil {
        return nil, errors.Wrap(err, "failed to filter logs")
    }

    return logs, nil
}

// logsToTransfers converts Ethereum logs to USDT transfers
func (l *Listener) logsToTransfers(ctx context.Context, logs []types.Log, blockNum uint64) ([]data.USDTTransfer, error) {
    transfers := make([]data.USDTTransfer, 0, len(logs))

    block, err := l.client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
    if err != nil {
        return nil, errors.Wrap(err, "failed to get block")
    }

    for _, log := range logs {
        transfer, err := l.logToTransfer(log, block.Time())
        if err != nil {
            l.log.WithError(err).WithField("txHash", log.TxHash.Hex()).Error("Failed to convert log to transfer")
            continue
        }
        transfers = append(transfers, transfer)
    }

    return transfers, nil
}

// logToTransfer converts a single Ethereum log to a USDT transfer
func (l *Listener) logToTransfer(log types.Log, blockTime uint64) (data.USDTTransfer, error) {
    if len(log.Topics) != 3 {
        return data.USDTTransfer{}, errors.New("invalid number of topics")
    }

    from := common.HexToAddress(log.Topics[1].Hex())
    to := common.HexToAddress(log.Topics[2].Hex())
    amount := new(big.Int).SetBytes(log.Data)

    return data.USDTTransfer{
        FromAddress:     from.Hex(),
        ToAddress:       to.Hex(),
        Amount:          amount.String(),
        TransactionHash: log.TxHash.Hex(),
        BlockNumber:     log.BlockNumber,
        LogIndex:        uint64(log.Index),
        Timestamp:       time.Unix(int64(blockTime), 0),
    }, nil
}

// Helper functions
func max(a, b uint64) uint64 {
    if a > b {
        return a
    }
    return b
}

func min(a, b uint64) uint64 {
    if a < b {
        return a
    }
    return b
}