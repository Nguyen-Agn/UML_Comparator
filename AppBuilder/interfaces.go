package AppBuilder

// Builder represents the top-level orchestrator that can execute a build process
type Builder interface {
	// Build executes the entire build pipeline for a specific project version.
	Build() error
}

// DirectoryPreparer isolates the responsibility of validating or creating output directories
type DirectoryPreparer interface {
	// Prepare ensures the target directory exists and is ready for output.
	Prepare(dirPath string) error
	// Clear removes all files and subdirectories from the target directory.
	Clear(dirPath string) error
}

// DependencyManager isolates the responsibility of managing module dependencies
type DependencyManager interface {
	// Tidy runs 'go mod tidy' to synchronize project dependencies.
	Tidy() error
}

// TaskBuilder isolates the responsibility of actually compiling binaries
type TaskBuilder interface {
	// BuildTask compiles a specific Go program with given flags and sources.
	BuildTask(name string, output string, sources []string, isGUI bool) error
}

// AssetCopier isolates the responsibility of moving physical file assets for packing
type AssetCopier interface {
	// CopyAssets copies files matching a filter from source to destination directory.
	CopyAssets(srcDir string, destDir string, filterExt string) error
}
