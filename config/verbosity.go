package config

import "github.com/opentable/sous/util/logging"

// Verbosity configures how chatty Sous is on its logs
type Verbosity struct {
	Silent, Quiet, Loud, Debug bool
}

// LoggingConfiguration produces a logging.Config equivalent to this simpler structure
func (v Verbosity) LoggingConfiguration() logging.Config {
	cfg := logging.Config{}
	switch {
	default:
		cfg.Basic.Level = "warning"
	case v.Silent:
		cfg.Basic.Level = "critical"
		cfg.Basic.DisableConsole = true
	case v.Quiet:
		cfg.Basic.Level = "critical"
	case v.Debug:
		cfg.Basic.Level = "debug"
	case v.Loud:
		cfg.Basic.Level = "extradebug1"
	}

	return cfg
}
