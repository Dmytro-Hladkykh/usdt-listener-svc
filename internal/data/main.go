package data

type MasterQ interface {
	New() MasterQ

	USDTTransfer() USDTTransferQ

	LastProcessedBlock() LastProcessedBlockQ

	Transaction(fn func(db MasterQ) error) error
}