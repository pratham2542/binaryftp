package main

import (
	binaryftpserver "binary-go/binaryftp/server"
	"io"
	"log"
	"os"
	"path/filepath"
)

const STORE_PATH = "./uploads/"

type FileStorage struct{}

func (FileStorage) Save(name string, r io.Reader, size uint64) error {

	path := filepath.Join(STORE_PATH, name)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.CopyN(file, r, int64(size))
	return err
}

func (FileStorage) Get(name string) (io.ReadCloser, uint64, error) {

	path := filepath.Join(STORE_PATH, name)

	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}

	return file, uint64(info.Size()), nil
}

func (FileStorage) List() ([]string, error) {

	files, err := os.ReadDir(STORE_PATH)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(files))

	for _, f := range files {
		if !f.IsDir() {
			names = append(names, f.Name())
		}
	}

	return names, nil
}

func main() {

	if err := os.MkdirAll(STORE_PATH, 0755); err != nil {
		log.Fatalf("could not create upload dir: %v", err)
	}

	storage := FileStorage{}

	server := binaryftpserver.New(":9000", storage)

	log.Fatal(server.Start())
}
