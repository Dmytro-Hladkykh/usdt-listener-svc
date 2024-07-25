package pg

import (
	"database/sql"

	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
	sq "github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const lastProcessedBlockTableName = "last_processed_block"

func NewLastProcessedBlockQ(db *pgdb.DB) data.LastProcessedBlockQ {
	return &lastProcessedBlockQ{
		db: db,
	}
}

type lastProcessedBlockQ struct {
	db *pgdb.DB
}

func (q *lastProcessedBlockQ) New() data.LastProcessedBlockQ {
	return NewLastProcessedBlockQ(q.db)
}

func (q *lastProcessedBlockQ) Get() (uint64, error) {
	var result uint64
	err := q.db.Get(&result, sq.Select("block_number").From(lastProcessedBlockTableName))
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to get last processed block from db")
	}
	return result, nil
}

func (q *lastProcessedBlockQ) Update(blockNumber uint64) error {
    query := sq.Update(lastProcessedBlockTableName).
        Set("block_number", blockNumber).
        Where(sq.Eq{"id": 1})

    err := q.db.Exec(query)
    if err != nil {
        return errors.Wrap(err, "failed to update last processed block in db")
    }
    return nil
}