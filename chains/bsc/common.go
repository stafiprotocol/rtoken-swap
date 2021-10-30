package bsc

import (
	"errors"

	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
)

var ErrNoOutPuts = errors.New("outputs length is zero")

func (w *writer) printContentError(m *core.Message, err error) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason, "err", err)
}

// submitMessage inserts the chainId into the msg and sends it to the router
func (w *writer) submitWriteMessage(m *core.Message) bool {
	err := w.router.SendWriteMesage(m)
	if err != nil {
		w.log.Error("failed to SendWriteMesage", "err", err)
		return false
	}
	return true
}

func (w *writer) submitReadMessage(m *core.Message) bool {
	err := w.router.SendReadMesage(m)
	if err != nil {
		w.log.Error("failed to SendReadMesage", "err", err)
		return false
	}
	return true
}

func (w *writer) reportTransResultWithBlock(source, dest core.RSymbol, m *submodel.TransResultWithBlock) bool {
	msg := &core.Message{Source: source, Destination: dest, Reason: core.ReportTransResultWithBlock, Content: m}
	return w.submitWriteMessage(msg)
}
