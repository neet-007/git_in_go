package utils

import (
	"os"
	"path/filepath"
)

func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.Mode().IsRegular(), nil
}

func RealPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err := os.Readlink(absPath)
		if err != nil {
			return "", err
		}
		return resolvedPath, nil
	}

	return absPath, nil
}
