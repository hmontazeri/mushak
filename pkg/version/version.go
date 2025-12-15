package version

// Version is the current version of mushak
// This will be set during build time via ldflags
var Version = "dev"

// GetVersion returns the current version
func GetVersion() string {
	return Version
}
