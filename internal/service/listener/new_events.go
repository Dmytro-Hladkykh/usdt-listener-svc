package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

// listenForNewEvents listens for new USDT events starting from the given block
func (l *Listener) listenForNewEvents(ctx context.Context, startBlock uint64) error {
	for {
		// Get the current block number
		currentBlock, err := l.client.BlockNumber(ctx)
		if err != nil {
			l.log.WithError(err).Error("Failed to get current block number")
			time.Sleep(ReconnectDelay)
			continue
		}

		// Process missed blocks
		for blockNum := startBlock; blockNum <= currentBlock; blockNum++ {
			if err := l.processBlock(ctx, blockNum); err != nil {
				l.log.WithError(err).WithField("blockNumber", blockNum).Error("Failed to process block")
				continue
			}
			startBlock = blockNum + 1
		}

		// Check if all blocks have been processed
		if startBlock > currentBlock {
			// Subscribe to new logs starting from the last processed block
			if err := l.subscribeToNewLogs(ctx, startBlock); err != nil {
				l.log.WithError(err).Error("Error in log subscription")
				time.Sleep(ReconnectDelay)
				continue
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue the loop
		}
	}
}

// subscribeToNewLogs subscribes to new USDT contract logs and processes them
func (l *Listener) subscribeToNewLogs(ctx context.Context, fromBlock uint64) error {
	// Set up the filter query for USDT contract logs
	contractAddress := common.HexToAddress(USDTContractAddress)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		Addresses: []common.Address{contractAddress},
	}

	// Subscribe to filtered logs
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
			// Process the new log
			tx, err := l.extractTransaction(vLog)
			if err != nil {
				l.log.WithError(err).Error("Error processing log")
				continue
			}

			// Insert the transaction into the database
			if _, err := l.db.USDTTransfer().Insert(tx); err != nil {
				l.log.WithError(err).Error("Failed to insert transaction")
			} else {
				// Log the processed transfer
				l.log.WithFields(logan.F{
					"from":      tx.FromAddress,
					"to":        tx.ToAddress,
					"amount":    tx.Amount,
					"txHash":    tx.TransactionHash,
					"timestamp": tx.Timestamp,
				}).Info("New USDT transfer processed")
			}

			// Update the last processed block
			if err := l.db.LastProcessedBlock().Update(vLog.BlockNumber); err != nil {
				l.log.WithError(err).Error("Failed to update last processed block")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}