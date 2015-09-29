package version

import (
	"sort"

	"github.com/opentable/sous/tools"
	"github.com/wmark/semver"
)

type V struct {
	Version  *semver.Version
	Original string
}

type R struct {
	Range    *semver.Range
	Original string
}

func Version(s string) *V {
	v, err := semver.NewVersion(s)
	if err != nil {
		tools.Dief("unable to parse version string '%s'; %s", s, err)
	}
	return &V{v, s}
}

type VL []*V

func VersionList(vs ...string) VL {
	list := make([]*V, len(vs))
	for i, v := range vs {
		list[i] = Version(v)
	}
	return list
}

func (l VL) Strings() []string {
	s := make([]string, len(l))
	for i, v := range l {
		s[i] = v.String()
	}
	return s
}

func Range(s string) *R {
	r, err := semver.NewRange(s)
	if err != nil {
		tools.Dief("unable to parse version range string '%s'; %s", s, err)
	}
	return &R{r, s}
}

type asc []*V

func (a asc) Len() int           { return len(a) }
func (a asc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a asc) Less(i, j int) bool { return a[i].Version.Less(a[j].Version) }

func (r *R) BestMatchFrom(versions []*V) *V {
	// Sort descending so we pick the highest compatible version
	sort.Reverse(asc(versions))
	for _, v := range versions {
		if r.Range.IsSatisfiedBy(v.Version) {
			return v
		}
	}
	return nil
}

func (r *R) IsSatisfiedBy(v *V) bool {
	return r.Range.IsSatisfiedBy(v.Version)
}

func (v *V) String() string {
	return v.Original
}
