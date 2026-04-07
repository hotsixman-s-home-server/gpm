package util

import (
	"os"
	"path/filepath"
)

func GetHomeDirPath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	homeDir := filepath.Join(dir, ".geep")

	return homeDir, nil
}

func GetUDSPath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "geep.sock"), nil
}
