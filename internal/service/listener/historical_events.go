package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

// processHistoricalEvents processes historical USDT events from the last processed block to the current block
func (l *Listener) processHistoricalEvents(ctx context.Context) (uint64, error) {
    // Get the last processed block
    lastProcessedBlock, err := l.db.LastProcessedBlock().Get()
    if err != nil {
        return 0, errors.Wrap(err, "failed to get last processed block")
    }

    for {
        // Get the current block number
        currentBlock, err := l.client.BlockNumber(ctx)
        if err != nil {
            return lastProcessedBlock, errors.Wrap(err, "failed to get current block number")
        }

        // Exit if all blocks have been processed
        if lastProcessedBlock >= currentBlock {
            break
        }

        // Log the range of blocks being processed
        l.log.WithFields(logan.F{
            "lastProcessedBlock": lastProcessedBlock,
            "currentBlock":       currentBlock,
        }).Info("Processing historical blocks")

        // Process blocks in batches
        batchSize := uint64(10)
        for blockNum := lastProcessedBlock + 1; blockNum <= currentBlock; blockNum += batchSize {
            // Calculate the end block for this batch
            endBlock := blockNum + batchSize - 1
            if endBlock > currentBlock {
                endBlock = currentBlock
            }

            // Process the block range
            if err := l.processBlockRange(ctx, blockNum, endBlock); err != nil {
                l.log.WithError(err).Error("Failed to process block range")
                time.Sleep(ReconnectDelay)
                continue
            }

            // Update the last processed block
            lastProcessedBlock = endBlock
            if err := l.db.LastProcessedBlock().Update(lastProcessedBlock); err != nil {
                l.log.WithError(err).Error("Failed to update last processed block")
            }
        }
    }

    return lastProcessedBlock, nil
}

// processBlockRange processes a range of blocks and extracts USDT transfers
func (l *Listener) processBlockRange(ctx context.Context, startBlock, endBlock uint64) error {
    for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
        // Process each block in the range
        if err := l.processBlock(ctx, blockNum); err != nil {
            l.log.WithError(err).WithField("blockNumber", blockNum).Error("Failed to process block")
            continue
        }

        // Update the last processed block in a transaction
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

// processBlock processes a single block and extracts USDT transfers
func (l *Listener) processBlock(ctx context.Context, blockNum uint64) error {
    // Set up the filter query for USDT contract logs
    contractAddress := common.HexToAddress(USDTContractAddress)
    query := ethereum.FilterQuery{
        FromBlock: big.NewInt(int64(blockNum)),
        ToBlock:   big.NewInt(int64(blockNum)),
        Addresses: []common.Address{contractAddress},
    }

    // Get logs from the block
    logs, err := l.client.FilterLogs(ctx, query)
    if err != nil {
        return errors.Wrap(err, "failed to filter logs")
    }

    // Extract transactions from logs
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

    // Insert transactions into the database
    if len(transactions) > 0 {
        if err := l.db.USDTTransfer().InsertBlock(transactions); err != nil {
            return errors.Wrap(err, "failed to insert block of USDT transfers")
        }
    }

    return nil
}