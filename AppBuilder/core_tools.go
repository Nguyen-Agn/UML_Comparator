package AppBuilder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StandardDirManager implements DirectoryPreparer
type StandardDirManager struct{}

func (m *StandardDirManager) Prepare(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("📁 Creating directory: %s\n", dirPath)
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

func (m *StandardDirManager) Clear(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err == nil {
		for _, f := range files {
			os.RemoveAll(filepath.Join(dirPath, f.Name()))
		}
	} else if os.IsNotExist(err) {
		return m.Prepare(dirPath)
	}
	return err
}

// GoDependencyManager implements DependencyManager
type GoDependencyManager struct{}

func (m *GoDependencyManager) Tidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GoTaskBuilder implements TaskBuilder
type GoTaskBuilder struct{}

func (b *GoTaskBuilder) BuildTask(name, output string, sources []string, isGUI bool) error {
	args := []string{"build"}
	if isGUI {
		args = append(args, "-ldflags=-H windowsgui")
	}
	args = append(args, "-o", output)
	args = append(args, sources...)

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FileAssetCopier implements AssetCopier
type FileAssetCopier struct{}

func (c *FileAssetCopier) CopyAssets(srcDir string, destDir, extensions string) error {
	paths := strings.Split(srcDir, ",")
	extList := strings.Split(extensions, ",")
	for i, e := range extList {
		extList[i] = strings.TrimSpace(e)
	}

	isMatch := func(name string) bool {
		if extensions == "" {
			return true
		}
		ext := filepath.Ext(name)
		for _, e := range extList {
			if e != "" && ext == e {
				return true
			}
		}
		return false
	}

	copied := 0

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if len(p) > 2 && p[0] == '"' && p[len(p)-1] == '"' {
			p = p[1 : len(p)-1]
		}
		if p == "" {
			continue
		}

		info, err := os.Stat(p)
		if err != nil {
			fmt.Printf("  ⚠️ Skipping invalid path: %v\n", err)
			continue
		}

		if !info.IsDir() {
			if isMatch(info.Name()) {
				dest := filepath.Join(destDir, info.Name())
				err := c.copyFile(p, dest)
				if err != nil {
					return err
				}
				fmt.Printf("  📄 Embedded: %s\n", info.Name())
				copied++
			}
		} else {
			entries, err := os.ReadDir(p)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					if isMatch(entry.Name()) {
						src := filepath.Join(p, entry.Name())
						dest := filepath.Join(destDir, entry.Name())
						err := c.copyFile(src, dest)
						if err != nil {
							return err
						}
						fmt.Printf("  📄 Embedded: %s\n", entry.Name())
						copied++
					}
				}
			}
		}
	}

	if copied == 0 {
		return fmt.Errorf("no files matching '%s' found in the provided paths", extensions)
	}
	return nil
}

func (c *FileAssetCopier) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}
