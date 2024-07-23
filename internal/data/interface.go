package data

import (
	"time"
)

type USDTTransfer struct {
    ID              int64     `db:"id"`
    FromAddress     string    `db:"from_address"`
    ToAddress       string    `db:"to_address"`
    Amount          string    `db:"amount"`
    TransactionHash string    `db:"transaction_hash"`
    BlockNumber     uint64    `db:"block_number"`
    Timestamp       time.Time `db:"timestamp"`
}

type USDTTransferQ interface {
    New() USDTTransferQ

    Get() (*USDTTransfer, error)
    Select() ([]USDTTransfer, error)
    Insert(transfer USDTTransfer) (*USDTTransfer, error)
    Update(transfer USDTTransfer) (*USDTTransfer, error)
    Delete() error

    FilterByID(id int64) USDTTransferQ
    FilterByFromAddress(address string) USDTTransferQ
    FilterByToAddress(address string) USDTTransferQ
    FilterByBlockNumber(blockNumber uint64) USDTTransferQ
    FilterByTransactionHash(hash string) USDTTransferQ
    
    OrderByTimestamp(desc bool) USDTTransferQ
    Limit(limit uint64) USDTTransferQ
    Offset(offset uint64) USDTTransferQ
}
