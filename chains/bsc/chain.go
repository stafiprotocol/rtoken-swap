package bsc

import (
	"errors"
	"fmt"

	"github.com/ChainSafe/log15"

	"rtoken-swap/core"
)

var ErrTerminated = errors.New("terminated")

type Chain struct {
	cfg    *core.ChainConfig // The config of the chain
	conn   *Connection
	writer *writer // The writer of the chain
	stop   chan<- int
}

func (c *Chain) InitBondedPools(symbols []core.RSymbol) error {
	return nil
}

func InitializeChain(cfg *core.ChainConfig, logger log15.Logger, sysErr chan<- error) (*Chain, error) {
	stop := make(chan int)
	conn, err := NewConnection(cfg, logger, stop)
	if err != nil {
		return nil, fmt.Errorf("cosmos NewConnection err: %s", err)
	}

	// Setup listener & writer
	w := NewWriter(conn, logger, sysErr)
	return &Chain{cfg: cfg, conn: conn, writer: w, stop: stop}, nil
}

func (c *Chain) Start() error {
	return c.writer.Start()
}

func (c *Chain) SetRouter(r *core.Router) {
	r.Listen(c.Rsymbol(), c.writer)
	c.writer.setRouter(r)
}

func (c *Chain) Rsymbol() core.RSymbol {
	return c.cfg.Symbol
}

func (c *Chain) Name() string {
	return c.cfg.Name
}

func (c *Chain) Stop() {
	close(c.stop)
}
