package main

import (
	"fmt"
	"log"
	"os"

	binaryftp "binary-go/binaryftp/client"
)

const (
	DOWNLOAD_PATH = "./downloads/"
)

func main() {
	client := binaryftp.New("localhost:9000")
	if err := os.MkdirAll(DOWNLOAD_PATH, 0755); err != nil {
		log.Fatalf("could not create upload dir: %v", err)
	}

	switch os.Args[1] {
	case "upload":
		err := client.Upload(os.Args[2])
		if err != nil {
			log.Fatal("Upload failed:", err)
		}
		fmt.Println("Upload successful")
	case "download":
		err := client.Download(os.Args[2], DOWNLOAD_PATH+os.Args[2])
		if err != nil {
			log.Fatal("Download failed:", err)
		}
		fmt.Println("Download successful")
	case "list":
		files, err := client.ListFiles()
		if err != nil {
			log.Fatal("List failed:", err)
		}
		fmt.Println("Files on server:")
		for _, f := range files {
			fmt.Println("-", f)
		}
	}
}
