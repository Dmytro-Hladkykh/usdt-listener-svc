package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Ethereum struct {
    RPCURL string `fig:"rpc_url,required"`
    StartingBlock uint64 `fig:"starting_block"`
}

type Ethereumer interface {
    Ethereum() *Ethereum
}

func NewEthereumer(getter kv.Getter) Ethereumer {
    return &ethereumConfig{
        getter: getter,
    }
}

type ethereumConfig struct {
    getter kv.Getter
    once   comfig.Once
}

func (e *ethereumConfig) Ethereum() *Ethereum {
    return e.once.Do(func() interface{} {
        var cfg Ethereum
        err := figure.Out(&cfg).From(kv.MustGetStringMap(e.getter, "ethereum")).Please()
        if err != nil {
            panic(errors.Wrap(err, "failed to figure out ethereum"))
        }
        return &cfg
    }).(*Ethereum)
}