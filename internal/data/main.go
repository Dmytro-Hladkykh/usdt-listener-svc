package data

type MasterQ interface {
	New() MasterQ

	USDTTransfer() USDTTransferQ

	Transaction(fn func(db MasterQ) error) error
}