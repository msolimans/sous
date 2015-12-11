package deploy

import (
	"fmt"
	"net/http"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/singularity"
)

type Diff struct {
	Desc       string
	Resolution func(s *singularity.Client) *http.Request
	Error      error
}

func ErrorDiff(err error) Diff {
	return Diff{Error: err}
}

func RequestMissingDiff(requestName string) Diff {
	return Diff{
		Desc: fmt.Sprintf("request %q does not exist", requestName),
	}
}

func (s *MergedState) Diff(dcName string) []Diff {
	dc := s.CompiledDatacentre(dcName)
	c := singularity.NewClient(dc.SingularityURL)
	rs, err := c.Requests()
	if err != nil {
		cli.Fatalf("%s", err)
	}
	cli.Logf("%s: %d", dc.SingularityURL, len(rs))
	return dc.DiffRequests()
}

func (d CompiledDatacentre) DiffRequests() []Diff {
	diffs := []Diff{}
	for _, m := range d.Manifests {
		diffs = append(diffs, m.Diff(d.SingularityURL)...)
	}
	return diffs
}

func (d DatacentreManifest) Diff(singularityURL string) []Diff {
	s := singularity.NewClient(singularityURL)
	r, err := s.Request(d.App.SourceRepo)
	if err != nil {
		return []Diff{ErrorDiff(err)}
	}
	if r == nil {
		return []Diff{RequestMissingDiff(d.App.SourceRepo)}
	}
	return nil
}
