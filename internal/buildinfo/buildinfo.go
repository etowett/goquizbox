// Package buildinfo provides high-level build information injected during
// build.
package buildinfo

var (
	// BuildID is the unique build identifier.
	BuildID string = "unknown"

	// BuildTag is the git tag from which this build was created.
	BuildTag string = "unknown"
)

// info provides the build information about the key server.
type buildinfo struct{}

// ID returns the build ID.
func (buildinfo) ID() string {
	return BuildID
}

// Tag returns the build tag.
func (buildinfo) Tag() string {
	return BuildTag
}
