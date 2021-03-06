package docker

import (
	"errors"
	"fmt"
	"io"

	sous "github.com/opentable/sous/lib"
	"github.com/opentable/sous/util/docker_registry"
	"github.com/opentable/sous/util/logging"
)

type selector struct {
	regClient docker_registry.Client
	log       logging.LogSink
}

// NewBuildStrategySelector constructs a sous.Selector that uses docker build images as its strategies
func NewBuildStrategySelector(ls logging.LogSink, rc docker_registry.Client) sous.Selector {
	return &selector{regClient: rc, log: ls}
}

// SelectBuildpack tries to select a buildpack for this BuildContext.
func (s *selector) SelectBuildpack(ctx *sous.BuildContext) (sous.Buildpack, error) {
	sbp := NewSplitBuildpack(s.regClient)
	dr, err := sbp.Detect(ctx)
	if err == nil && dr.Compatible {
		reportStrategyChoice("split container", s.log)
		return sbp, nil
	}

	dfbp := NewDockerfileBuildpack()
	dr, err = dfbp.Detect(ctx)
	if err == nil && dr.Compatible {
		reportStrategyChoice("simple dockerfile", s.log)
		return dfbp, nil
	}
	return nil, errors.New("no buildpack detected for project")
}

type strategyChoiceMessage struct {
	logging.CallerInfo
	choice string
}

func reportStrategyChoice(choice string, log logging.LogSink) {
	msg := strategyChoiceMessage{
		choice:     choice,
		CallerInfo: logging.GetCallerInfo(logging.NotHere()),
	}
	logging.Deliver(msg, log)
}

func (msg strategyChoiceMessage) WriteToConsole(console io.Writer) {
	fmt.Fprintf(console, "Building with %s\n", msg.choice)
}

func (msg strategyChoiceMessage) DefaultLevel() logging.Level {
	return logging.DebugLevel
}

func (msg strategyChoiceMessage) Message() string {
	return msg.choice
}
