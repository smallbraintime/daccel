package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	MB10  = 10000000
	MB100 = 100000000
	GB1   = 1000000000
)

func downloadFile(url string, chunkSize int, fileSize int, file *os.File) {
	mtx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	transport := &http.Transport{
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
	}

	for startByte := 0; startByte < fileSize; startByte += chunkSize {
		endByte := startByte + chunkSize - 1
		if endByte >= fileSize {
			endByte = fileSize - 1
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			reader, err := NewChunkDownloader(url, start, end, client)
			if err != nil {
				panic(err)
			}
			defer reader.Close()

			writer := NewFileChunkWriter(file, mtx, start)

			io.Copy(writer, reader)
		}(startByte, endByte)
	}

	wg.Wait()
}

func download(url string, destFile *os.File) error {
	resp, err := http.Head(url)
	var fileSize int
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fileSize, err = strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	var chunkSize int
	switch {
	case fileSize <= MB10:
		chunkSize = fileSize
	case fileSize <= MB100:
		chunkSize = fileSize / 2
	case fileSize <= GB1:
		chunkSize = fileSize / 4
	default:
		chunkSize = fileSize / 8
	}

	downloadFile(url, chunkSize, fileSize, destFile)

	return nil
}

func createFile(sourceUrl, destPath string) *os.File {
	u, err := url.Parse(sourceUrl)
	if err != nil {
		log.Fatal(err)
	}

	fileName := filepath.Base(u.Path)

	var file *os.File
	if destPath == "" {
		destPath, err = os.Getwd()
		file, err = os.Create(destPath + "/" + fileName)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		_, err := os.Stat(destPath)
		if err != nil {
			fmt.Println(err)
		}
		file, err = os.Create(destPath + "/" + fileName)
		if err != nil {
			fmt.Println(err)
		}
	}

	return file
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("an error occured while downloading")
		}
	}()

	sourceUrl := ""
	destPath := ""

	args := os.Args
	switch len(args) {
	case 1:
		fmt.Println("daccel 0.1")
		fmt.Println("Usage: daccel <url> [dest]")
		return
	case 2:
		sourceUrl = args[1]
	case 3:
		sourceUrl = args[1]
		destPath = args[2]
	default:
		fmt.Println(errors.New("invalid arguments"))
	}

	file := createFile(sourceUrl, destPath)
	defer file.Close()

	now := time.Now()
	start := now.Unix()

	fmt.Println("downloading...")
	err := download(sourceUrl, file)

	now = time.Now()
	stop := now.Unix()

	fmt.Println("finished in downloading in ", stop-start, " seconds")

	if err != nil {
		fmt.Println("failed to connect")
	}
}
