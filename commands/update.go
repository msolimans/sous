package commands

import (
	"time"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/config"
	"github.com/opentable/sous/tools/cli"
)

func UpdateHelp() string {
	return `sous update updates your local sous config cache`
}

func Update(packs []*build.Pack, args []string) {
	key := "last-update-check"
	if err := config.Update(); err != nil {
		cli.Fatal()
	}
	config.Set(key, time.Now().Format(time.RFC3339))
	cli.Success()
}
