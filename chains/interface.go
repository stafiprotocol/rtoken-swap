package chains

// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

import (
	"rtoken-swap/core"
)

type Router interface {
	SendWriteMesage(msg *core.Message) error
	SendReadMesage(msg *core.Message) error
}
