package pg

import (
	"github.com/Dmytro-Hladkykh/usdt-listener-svc/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

func NewMasterQ(db *pgdb.DB) data.MasterQ {
    return &masterQ{
        db: db.Clone(),
    }
}

type masterQ struct {
    db *pgdb.DB
}

func (m *masterQ) New() data.MasterQ {
    return NewMasterQ(m.db)
}

func (m *masterQ) USDTTransfer() data.USDTTransferQ {
    return NewUSDTTransferQ(m.db)
}

func (m *masterQ) LastProcessedBlock() data.LastProcessedBlockQ {
	return NewLastProcessedBlockQ(m.db)
}

func (m *masterQ) Transaction(fn func(q data.MasterQ) error) error {
    return m.db.Transaction(func() error {
        return fn(m)
    })
}