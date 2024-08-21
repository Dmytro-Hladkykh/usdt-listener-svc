package config

import (
	"fmt"

	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Ethereum struct {
    RPCURL        string `fig:"rpc_url,required"`
    StartingBlock uint64 `fig:"starting_block,required"`
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
        
        raw := kv.MustGetStringMap(e.getter, "ethereum")
        
        err := figure.Out(&cfg).From(raw).Please()
        if err != nil {
            fmt.Printf("Error figuring out ethereum config: %v\n", err)
            panic(errors.Wrap(err, "failed to figure out ethereum config"))
        }
                
        // Validate the configuration
        if cfg.RPCURL == "" {
            panic(errors.New("ethereum RPC URL is not set"))
        }
        if cfg.StartingBlock == 0 {
            panic(errors.New("ethereum starting block is not set"))
        }
        
        return &cfg
    }).(*Ethereum)
}