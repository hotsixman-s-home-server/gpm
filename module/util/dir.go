package util

import (
	"os"
	"path/filepath"
)

func GetGeepDir() (string, error) {
	envGeepDir := os.Getenv("GEEP_DIR")
	if envGeepDir != "" {
		return envGeepDir, nil
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	geepDir := filepath.Join(dir, ".geep")

	return geepDir, nil
}

func GetUDSPath() (string, error) {
	dir, err := GetGeepDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "geep.sock"), nil
}
