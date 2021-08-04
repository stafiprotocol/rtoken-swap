package substrate

import (
	"fmt"
	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
)

type eventName string
type eventHandler func(interface{}) (*core.Message, error)

type eventHandlerSubscriptions struct {
	name    eventName
	handler eventHandler
}

var (
	NewMultisig      = eventName(config.NewMultisigEventId)
	MultisigExecuted = eventName(config.MultisigExecutedEventId)
)

var Subscriptions = []eventHandlerSubscriptions{
	{NewMultisig, newMultisigHandler},
	{MultisigExecuted, multisigExecutedHandler},
}

func newMultisigHandler(data interface{}) (*core.Message, error) {
	d, ok := data.(*submodel.EventNewMultisig)
	if !ok {
		return nil, fmt.Errorf("failed to cast newMultisig")
	}

	return &core.Message{Reason: core.NewMultisig, Content: d}, nil
}

func multisigExecutedHandler(data interface{}) (*core.Message, error) {
	d, ok := data.(*submodel.EventMultisigExecuted)
	if !ok {
		return nil, fmt.Errorf("failed to cast multisigExecuted")
	}

	return &core.Message{Reason: core.MultisigExecuted, Content: d}, nil
}
