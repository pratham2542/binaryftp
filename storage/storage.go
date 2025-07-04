package storage

import (
	"os"
	"path/filepath"
)

const StoreDir = "./ftp_data"

func init() {
	_ = os.MkdirAll(StoreDir, 0755)
}

func SaveFile(filename string, data []byte) error {
	path := filepath.Join(StoreDir, filepath.Base(filename))
	return os.WriteFile(path, data, 0644)
}

func LoadFile(filename string) ([]byte, error) {
	path := filepath.Join(StoreDir, filepath.Base(filename))
	return os.ReadFile(path)
}

func ListFiles() ([]string, error) {
	entries, err := os.ReadDir(StoreDir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files, nil
}
