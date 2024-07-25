package pg

import (
	"database/sql"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
	sq "github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const usdtTransfersTableName = "usdt_transfers"

func NewUSDTTransferQ(db *pgdb.DB) data.USDTTransferQ {
	return &usdtTransferQ{
		db:  db,
		sql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From(usdtTransfersTableName),
	}
}

type usdtTransferQ struct {
	db  *pgdb.DB
	sql sq.SelectBuilder
}

func (q *usdtTransferQ) New() data.USDTTransferQ {
	return NewUSDTTransferQ(q.db)
}

func (q *usdtTransferQ) Get() (*data.USDTTransfer, error) {
	var result data.USDTTransfer
	stmt := q.sql
	err := q.db.Get(&result, stmt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get USDT transfer from db")
	}
	return &result, nil
}

func (q *usdtTransferQ) Select() ([]data.USDTTransfer, error) {
	var result []data.USDTTransfer
	stmt := q.sql
	err := q.db.Select(&result, stmt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to select USDT transfers from db")
	}
	return result, nil
}

func (q *usdtTransferQ) Insert(transfer data.USDTTransfer) (*data.USDTTransfer, error) {
	clauses := map[string]interface{}{
		"from_address":     transfer.FromAddress,
		"to_address":       transfer.ToAddress,
		"amount":           transfer.Amount,
		"transaction_hash": transfer.TransactionHash,
		"block_number":     transfer.BlockNumber,
		"log_index":        transfer.LogIndex,
		"timestamp":        transfer.Timestamp,
	}
	var result data.USDTTransfer
	stmt := sq.Insert(usdtTransfersTableName).SetMap(clauses).Suffix("RETURNING *")
	err := q.db.Get(&result, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert USDT transfer to db")
	}
	return &result, nil
}

func (q *usdtTransferQ) InsertIgnore(transfer data.USDTTransfer) (*data.USDTTransfer, error) {
    clauses := map[string]interface{}{
        "from_address":     transfer.FromAddress,
        "to_address":       transfer.ToAddress,
        "amount":           transfer.Amount,
        "transaction_hash": transfer.TransactionHash,
        "block_number":     transfer.BlockNumber,
        "log_index":        transfer.LogIndex,
        "timestamp":        transfer.Timestamp,
    }
    var result data.USDTTransfer
    stmt := sq.Insert(usdtTransfersTableName).SetMap(clauses).Suffix("ON CONFLICT (transaction_hash, log_index) DO NOTHING RETURNING *")
    err := q.db.Get(&result, stmt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, errors.Wrap(err, "failed to insert USDT transfer to db")
    }
    return &result, nil
}

func (q *usdtTransferQ) InsertBlock(transfers []data.USDTTransfer) error {
    if len(transfers) == 0 {
        return nil
    }

    insert := sq.Insert(usdtTransfersTableName).Columns(
        "from_address", "to_address", "amount", "transaction_hash", 
        "block_number", "log_index", "timestamp",
    )

    for _, transfer := range transfers {
        insert = insert.Values(
            transfer.FromAddress, transfer.ToAddress, transfer.Amount,
            transfer.TransactionHash, transfer.BlockNumber, transfer.LogIndex,
            transfer.Timestamp,
        )
    }

    insert = insert.Suffix("ON CONFLICT (transaction_hash, log_index) DO NOTHING")

    err := q.db.Exec(insert)
    return errors.Wrap(err, "failed to insert batch of USDT transfers")
}


func (q *usdtTransferQ) Update(transfer data.USDTTransfer) (*data.USDTTransfer, error) {
	clauses := map[string]interface{}{
		"from_address":     transfer.FromAddress,
		"to_address":       transfer.ToAddress,
		"amount":           transfer.Amount,
		"transaction_hash": transfer.TransactionHash,
		"block_number":     transfer.BlockNumber,
		"log_index":        transfer.LogIndex,
		"timestamp":        transfer.Timestamp,
	}
	var result data.USDTTransfer
	stmt := sq.Update(usdtTransfersTableName).SetMap(clauses).Where(sq.Eq{"id": transfer.ID}).Suffix("RETURNING *")
	err := q.db.Get(&result, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update USDT transfer in db")
	}
	return &result, nil
}

func (q *usdtTransferQ) FilterByID(id int64) data.USDTTransferQ {
	q.sql = q.sql.Where(sq.Eq{"id": id})
	return q
}

func (q *usdtTransferQ) FilterByFromAddress(address string) data.USDTTransferQ {
	q.sql = q.sql.Where(sq.Eq{"from_address": address})
	return q
}

func (q *usdtTransferQ) FilterByToAddress(address string) data.USDTTransferQ {
	q.sql = q.sql.Where(sq.Eq{"to_address": address})
	return q
}

func (q *usdtTransferQ) FilterByBlockNumber(blockNumber uint64) data.USDTTransferQ {
	q.sql = q.sql.Where(sq.Eq{"block_number": blockNumber})
	return q
}

func (q *usdtTransferQ) FilterByTransactionHash(hash string) data.USDTTransferQ {
	q.sql = q.sql.Where(sq.Eq{"transaction_hash": hash})
	return q
}

func (q *usdtTransferQ) OrderByTimestamp(desc bool) data.USDTTransferQ {
	if desc {
		q.sql = q.sql.OrderBy("timestamp DESC")
	} else {
		q.sql = q.sql.OrderBy("timestamp ASC")
	}
	return q
}

func (q *usdtTransferQ) Limit(limit uint64) data.USDTTransferQ {
	q.sql = q.sql.Limit(limit)
	return q
}

func (q *usdtTransferQ) Offset(offset uint64) data.USDTTransferQ {
	q.sql = q.sql.Offset(offset)
	return q
}

func (q *usdtTransferQ) Page(pageParams *pgdb.OffsetPageParams) data.USDTTransferQ {
    q.sql = pageParams.ApplyTo(q.sql, "id")
    return q
}
