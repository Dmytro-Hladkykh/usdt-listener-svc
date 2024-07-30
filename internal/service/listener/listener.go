package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/config"
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
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
    config config.Config
}

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

    if processHist {
        startBlock, err = l.processHistoricalEvents(ctx, startBlock)
        if err != nil {
            l.log.WithError(err).Error("Failed to process historical events")
        }
        // Re-process the last few blocks to ensure no transactions are missed
        overlapBlockCount := uint64(1)
        if startBlock > overlapBlockCount {
            startBlock -= overlapBlockCount
        } else {
            startBlock = 0
        }
    } else {
        l.log.Info("Skipping historical events processing")
        if err := l.db.LastProcessedBlock().Update(startBlock); err != nil {
            l.log.WithError(err).Error("Failed to update last processed block")
        }
    }

    // Start listening for new events
    return l.listenForNewEvents(ctx, startBlock)
}

// extractTransaction extracts USDT transfer data from a log event
func (l *Listener) extractTransaction(vLog types.Log) (data.USDTTransfer, error) {
    // Check if the log has the correct number of topics
    if len(vLog.Topics) != 3 {
        return data.USDTTransfer{}, errors.New("invalid log topics length")
    }

    // Extract transfer details from log topics and data
    from := common.HexToAddress(vLog.Topics[1].Hex())
    to := common.HexToAddress(vLog.Topics[2].Hex())
    amount := new(big.Int).SetBytes(vLog.Data)

    block, err := l.client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
    if err != nil {
        return data.USDTTransfer{}, errors.Wrap(err, "failed to get block information")
    }

    // Create and return the USDT transfer object
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