package sous

import "github.com/nyarly/spies"

type (
	// ImageLabeller can get the image labels for a given imageName
	ImageLabeller interface {
		//ImageLabels finds the sous (docker) labels for a given image name
		ImageLabels(imageName string) (labels map[string]string, err error)
	}

	// Registry describes a system for mapping SourceIDs to BuildArtifacts and vice versa
	Registry interface {
		ImageLabeller
		// GetArtifact gets the build artifact address for a source ID.
		// It does not guarantee that that artifact exists.
		GetArtifact(SourceID) (*BuildArtifact, error)
		// GetSourceID gets the source ID associated with the
		// artifact, regardless of the existence of the artifact.
		GetSourceID(*BuildArtifact) (SourceID, error)
		// GetMetadata returns metadata for a source ID.
		//GetMetadata(SourceID) (map[string]string, error)

		// ListSourceIDs returns a list of known SourceIDs
		ListSourceIDs() ([]SourceID, error)

		// Warmup requests that the registry check specific artifact names for existence
		// the details of this behavior will vary by implementation. For Docker, for instance,
		// the corresponding repo is enumerated
		Warmup(string) error
	}

	// An Inserter puts data into a registry.
	Inserter interface {
		// Insert pairs a SourceID with an imagename, and tags the pairing with Qualities
		// The etag can be (usually will be) the empty string
		Insert(sid SourceID, in, etag string, qs []Quality) error
	}
)

type (
	// An InserterSpy is a spy implementation of the Inserter interface
	InserterSpy struct {
		*spies.Spy
	}
)

// NewInserterSpy returns a spy inserter for testing
func NewInserterSpy() InserterSpy {
	return InserterSpy{spies.NewSpy()}
}

// Insert implements Inserter on InserterSpy
func (is InserterSpy) Insert(sid SourceID, in, etag string, qs []Quality) error {
	return is.Called(sid, in, etag, qs).Error(0)
}
