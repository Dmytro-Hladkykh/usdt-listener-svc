package data

import (
	"time"

	"gitlab.com/distributed_lab/kit/pgdb"
)

type USDTTransfer struct {
    ID              int64     `db:"id"`
    FromAddress     string    `db:"from_address"`
    ToAddress       string    `db:"to_address"`
    Amount          string    `db:"amount"`
    TransactionHash string    `db:"transaction_hash"`
    BlockNumber     uint64    `db:"block_number"`
    LogIndex        uint64    `db:"log_index"`
    Timestamp       time.Time `db:"timestamp"`
}

type LastProcessedBlock struct {
    ID          int64  `db:"id"`
    BlockNumber uint64 `db:"block_number"`
}

type USDTTransferQ interface {
    New() USDTTransferQ

    Get() (*USDTTransfer, error)
    Select() ([]USDTTransfer, error)
    Insert(transfer USDTTransfer) (*USDTTransfer, error)
    InsertIgnore(transfer USDTTransfer) (*USDTTransfer, error)
    InsertBlock(transfer []USDTTransfer) error
    DeleteLastProcessedBlock(blockNumber uint64) error
    Update(transfer USDTTransfer) (*USDTTransfer, error)

    FilterByID(id int64) USDTTransferQ
    FilterByFromAddress(address string) USDTTransferQ
    FilterByToAddress(address string) USDTTransferQ
    FilterByBlockNumber(blockNumber uint64) USDTTransferQ
    FilterByTransactionHash(hash string) USDTTransferQ
    
    OrderByTimestamp(desc bool) USDTTransferQ
    Limit(limit uint64) USDTTransferQ
    Offset(offset uint64) USDTTransferQ

    Page(pageParams *pgdb.OffsetPageParams) USDTTransferQ
}

type LastProcessedBlockQ interface {
    New() LastProcessedBlockQ

    Get() (uint64, error)
    Update(blockNumber uint64) error
}