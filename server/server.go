package main

import (
	binaryftpserver "binary-go/binaryftp/server"
	"fmt"
	"log"
	"os"
)

const STORE_PATH = "./uploads/"

func main() {
	server := binaryftpserver.New(":9000")

	if err := os.MkdirAll(STORE_PATH, 0755); err != nil {
		log.Fatalf("could not create upload dir: %v", err)
	}

	server.OnUpload(func(filename string, data []byte) error {
		path := STORE_PATH + filename
		fmt.Println("Writing file to:", path)

		err := os.WriteFile(path, data, 0644)
		if err != nil {
			fmt.Println("Error saving file:", err)
		} else {
			fmt.Println("File saved successfully.")
		}
		return err
	})

	server.OnDownload(func(filename string) ([]byte, error) {
		path := STORE_PATH + filename
		file, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error saving file:", err)
			return nil, err
		} else {
			fmt.Println("File saved successfully.")
		}
		return file, nil
	})

	server.OnList(func() ([]string, error) {
		files, err := os.ReadDir(STORE_PATH)
		if err != nil {
			return nil, err
		}
		var names []string
		for _, f := range files {
			if !f.IsDir() {
				names = append(names, f.Name())
			}
		}
		return names, nil
	})

	log.Fatal(server.Start())
}
