package AppBuilder

// Builder represents the top-level orchestrator that can execute a build process
type Builder interface {
	Build() error
}

// DirectoryPreparer isolates the responsibility of validating or creating output directories
type DirectoryPreparer interface {
	Prepare(dirPath string) error
	Clear(dirPath string) error
}

// DependencyManager isolates the responsibility of managing module dependencies
type DependencyManager interface {
	Tidy() error
}

// TaskBuilder isolates the responsibility of actually compiling binaries
type TaskBuilder interface {
	BuildTask(name string, output string, sources []string, isGUI bool) error
}

// AssetCopier isolates the responsibility of moving physical file assets for packing
type AssetCopier interface {
	CopyAssets(srcDir string, destDir string, filterExt string) error
}
