// Package cli implements the Sous Command Line Interface. It is a
// presentation layer, and contains no core logic.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/graph"
	"github.com/opentable/sous/util/cmdr"
	"github.com/opentable/sous/util/logging"
	"github.com/opentable/sous/util/yaml"
	"github.com/pkg/errors"
)

// Func aliases, for convenience returning from commands.
var (
	GeneralErrorf = func(format string, a ...interface{}) cmdr.ErrorResult {
		return EnsureErrorResult(fmt.Errorf(format, a...))
	}
	EnsureErrorResult = func(err error) cmdr.ErrorResult {
		logging.Log.Debugf("%#v", err)
		return cmdr.EnsureErrorResult(err)
	}
)

// ProduceResult converts errors into Results
func ProduceResult(err error) cmdr.Result {
	if err != nil {
		return EnsureErrorResult(err)
	}

	return cmdr.Success()
}

type (
	// CLI describes the command line interface for Sous
	CLI struct {
		*cmdr.CLI
		graph *graph.SousGraph
	}
	// Addable objects are able to receive lists of interface{}, presumably to add
	// them to a DI registry. Abstracts Psyringe's Add()
	Addable interface {
		Add(...interface{})
	}
)

// SuccessYAML lets you return YAML on the command line.
func SuccessYAML(v interface{}) cmdr.Result {
	b, err := yaml.Marshal(v)
	if err != nil {
		return cmdr.InternalErrorf("unable to marshal YAML: %s", err)
	}
	return cmdr.SuccessData(b)
}

// Plumbing injects a command with the current psyringe,
// then it Executes it, returning the result.
func (cli *CLI) Plumbing(from cmdr.Command, cmd cmdr.Executor, args []string) cmdr.Result {
	if err := cli.Plumb(from, cmd); err != nil {
		return cmdr.EnsureErrorResult(err)
	}
	return cmd.Execute(args)
}

// Plumb injects a lists of commands with the current psyringe,
// returning early in the event of an error
func (cli *CLI) Plumb(from cmdr.Command, cmds ...cmdr.Executor) error {
	for _, cmd := range cmds {
		if err := cli.graph.Inject(cmd); err != nil {
			return err
		}
	}
	return nil
}

// buildCLIGraph builds the CLI DI graph.
func buildCLIGraph(root *Sous, cli *CLI, out, err io.Writer) {
	g := cli.graph
	g.Add(cli)
	g.Add(root)
	g.Add(func(c *CLI) graph.Out {
		return graph.Out{Output: c.Out}
	})
	g.Add(func(c *CLI) graph.ErrOut {
		return graph.ErrOut{Output: c.Err}
	})
}

// NewSousCLI creates a new Sous cli app.
func NewSousCLI(di *graph.SousGraph, s *Sous, out, errout io.Writer) (*CLI, error) {

	stdout := cmdr.NewOutput(out)
	stderr := cmdr.NewOutput(errout)

	var verbosity config.Verbosity

	cli := &CLI{
		CLI: &cmdr.CLI{
			Root: s,
			Out:  stdout,
			Err:  stderr,
			// HelpCommand is shown to the user if they type something that looks
			// like they want help, but which isn't recognised by Sous properly. It
			// uses the standard flag.ErrHelp value to decide whether or not to show
			// this.
			HelpCommand: os.Args[0] + " help",
			GlobalFlagSetFuncs: []func(*flag.FlagSet){
				func(fs *flag.FlagSet) {
					fs.BoolVar(&verbosity.Silent, "s", false,
						"silent: silence all non-essential output")
					fs.BoolVar(&verbosity.Quiet, "q", false,
						"quiet: output only essential error messages")
					fs.BoolVar(&verbosity.Loud, "v", false,
						"loud: output extra info, including all shell commands")
					fs.BoolVar(&verbosity.Debug, "d", false,
						"debug: output detailed logs of internal operations")
				},
			},
		},
		graph: di,
	}

	buildCLIGraph(s, cli, out, errout)

	addVerbosityOnce := sync.Once{}

	cli.Hooks.Parsed = func(cmd cmdr.Command) error {
		addVerbosityOnce.Do(func() {
			cli.graph.Add(&verbosity)
		})
		if registrant, ok := cmd.(interface {
			RegisterOn(Addable)
		}); ok {
			registrant.RegisterOn(cli.graph)
		}
		return nil
	}

	// Before Execute is called on any command, inject its dependencies.
	cli.Hooks.PreExecute = func(cmd cmdr.Command) error {
		return cli.graph.Inject(cmd)
	}

	cli.Hooks.PreFail = func(err error) cmdr.ErrorResult {
		if err != nil {
			originalErr := fmt.Sprint(err)
			err = errors.Cause(err)
			causeStr := err.Error()
			if originalErr != causeStr {
				logging.Log.Debugf("%v\n", originalErr)
			}
		}
		return EnsureErrorResult(err)
	}

	return cli, nil
}
