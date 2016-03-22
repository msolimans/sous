package cli

type SousVersion struct {
	Sous *Sous
}

const sousVersionHelp = `
print the version of sous

prints the current version of sous. Please include the output from this
command with any bug reports sent to https://github.com/opentable/sous/issues

args:

Sous is versioned using semver. There are three versioned pieces of Sous:
Sous Engine, Sous Server, and Sous CLI.
`

func (*SousVersion) Help() *Help {
	return ParseHelp(sousVersionHelp)
}

func (sv *SousVersion) Execute(args []string) Result {
	return Successf("sous version %s", sv.Sous.Version)
}