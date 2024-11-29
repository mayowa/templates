package templates

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (t *Template) Exists(name string) bool {
	var (
		found bool
	)

	t.mtx.RLock()
	_, found = t.cache[name]
	t.mtx.RUnlock()

	return found
}

// isFolder checks if a folder exists in the template folder
func (t *Template) isFolder(name string) bool {
	templateName := filepath.Join(t.root, name)
	fi, err := os.Stat(templateName)
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

func (t *Template) pathExists(name string) bool {
	_, err := os.Stat(name)

	return err == nil
}

func (t *Template) stripFileName(name string) string {
	if strings.HasPrefix(name, t.sharedFolder) {
		name = strings.TrimPrefix(name, t.sharedFolder)
	} else {
		name = strings.TrimPrefix(name, filepath.Clean(t.root))
	}

	cf := filepath.Join("/", t.componentFolder) + "/"
	if strings.HasPrefix(name, cf) {
		name = strings.TrimPrefix(name, cf)
	} else {
		name = strings.TrimSuffix(name, t.ext)
	}
	if name[0] == '/' {
		name = name[1:]
	}
	return name
}

func (t *Template) sortBlockFiles(blockName string, files []string) {
	// put the file with the same name as the block first
	idx := -1
	for i, fle := range files {
		fle, _ = filepath.Abs(fle)
		fle = strings.TrimSuffix(fle, t.ext)
		fle = filepath.Base(fle)
		if fle == blockName {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}
	files[0], files[idx] = files[idx], files[0]
}

// findFiles
func (t *Template) findFiles(root, ext string) (filenames []string, err error) {

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ext) {
			filenames = append(filenames, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(filenames) == 0 {
		return nil, fmt.Errorf("no file found in: %#q", root)
	}

	return filenames, nil
}
