package listener

import (
	"context"
	"math/big"
	"os"
	"time"

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
}

func NewListener(infuraURL string, db data.MasterQ, log *logan.Entry) (*Listener, error) {
    // Connect to the Ethereum client
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

// Listen starts the main loop for listening to USDT transfers
func (l *Listener) Listen(ctx context.Context) error {
    processHist := os.Getenv("PROCESS_HIST")

    var lastProcessedBlock uint64
    var err error

    // Process historical events if enabled
    if processHist == "true" {
        lastProcessedBlock, err = l.processHistoricalEvents(ctx)
        if err != nil {
            l.log.WithError(err).Error("Failed to process historical events")
        }
    } else {
        // Skip historical processing and start from the current block
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

    // Start listening for new events
    return l.listenForNewEvents(ctx, lastProcessedBlock)
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