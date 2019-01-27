package cli

import (
	"context"

	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/kscmd"
)

type Options struct {
	Url            string
	Json           bool
	DebugContainer DebugContainer
	// Debug Container is a superset of DebugRequest so we can use the same struct
	// TODO(mitchdraft) - refactor
	DebugRequest DebugContainer
	daClient     *v1.DebugAttachmentClient
	ctx          context.Context
	Wait         Wait
	LiteOptions  kscmd.SquashConfig
}

type DebugContainer struct {
	Name         string
	Namespace    string
	Image        string
	Pod          string
	Container    string
	ProcessName  string
	DebuggerType string
}

type Wait struct {
	Timeout float64
}

type Error struct {
	Type string
	Info string
}
