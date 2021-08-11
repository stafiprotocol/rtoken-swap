// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package substrate

import (
	"errors"
	"strconv"

	"rtoken-swap/config"
	"rtoken-swap/core"

	"github.com/ChainSafe/log15"
)

var ErrorTerminated = errors.New("terminated")

type Chain struct {
	cfg      *core.ChainConfig // The config of the chain
	conn     *Connection
	listener *listener // The listener of this chain
	writer   *writer   // The writer of the chain
	stop     chan<- int
}

func InitializeChain(cfg *core.ChainConfig, logger log15.Logger, sysErr chan<- error) (*Chain, error) {
	logger.Info("InitializeChain", "symbol", cfg.Symbol)

	stop := make(chan int)
	conn, err := NewConnection(cfg, logger, stop)
	if err != nil {
		return nil, err
	}

	useStartBlock := uint64(0)
	latestBlock, err := conn.LatestBlockNumber()
	if err != nil {
		return nil, err
	}
	startBlock := parseStartBlock(cfg)
	if cfg.LatestBlockFlag {
		useStartBlock = latestBlock
	} else {
		useStartBlock = startBlock
	}
	thresholdStr, ok := cfg.Opts[config.ThresholdKey].(string)
	if !ok {
		return nil, errors.New("config must has threshold")
	}
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		return nil, err
	}

	// Setup listener & writer
	l := NewListener(cfg.Name, cfg.Symbol, useStartBlock, conn, logger, stop, sysErr)
	w := NewReaderWriter(cfg.Symbol, cfg.Opts, conn, threshold, logger, sysErr, stop)
	return &Chain{cfg: cfg, conn: conn, listener: l, writer: w, stop: stop}, nil
}

func (c *Chain) Start() error {
	err := c.listener.start()
	if err != nil {
		return err
	}

	err = c.writer.Start()
	if err != nil {
		return err
	}
	return nil
}

func (c *Chain) SetRouter(r *core.Router) {
	r.Listen(c.Rsymbol(), c.writer)
	c.listener.setRouter(r)
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

func parseStartBlock(cfg *core.ChainConfig) uint64 {
	if blk, ok := cfg.Opts[config.StartBlockKey]; ok {
		blkStr, ok := blk.(string)
		if !ok {
			panic("block not string")
		}
		res, err := strconv.ParseUint(blkStr, 10, 32)
		if err != nil {
			panic(err)
		}
		return res
	}
	return 0
}
