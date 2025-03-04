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
)

const (
	MB10  = 10000000
	MB100 = 100000000
	GB1   = 1000000000
)

func NewChunkDownloader(url string, startByte int, endByte int) (*ChunkDownloader, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &ChunkDownloader{resp.Body, 0}, nil
}

type ChunkDownloader struct {
	reader      io.ReadCloser
	currentByte int
}

func (c *ChunkDownloader) Read(p []byte) (n int, err error) {
	bytesRead, err := c.reader.Read(p)
	if err != nil {
		if err == io.EOF {
			return bytesRead, nil
		}
		return bytesRead, err
	}

	if bytesRead > 0 {
		c.currentByte += bytesRead
	}

	fmt.Println("read: ", bytesRead, " bytes")
	return bytesRead, err
}

func (c *ChunkDownloader) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

func NewFileChunkWriter(file *os.File, mtx *sync.Mutex, startByte int) *FileChunkWriter {
	return &FileChunkWriter{file, mtx, startByte}
}

type FileChunkWriter struct {
	file        *os.File
	mtx         *sync.Mutex
	currentByte int
}

func (f *FileChunkWriter) Write(p []byte) (n int, err error) {
	fmt.Println("locked")
	f.mtx.Lock()
	defer f.mtx.Unlock()

	_, err = f.file.Seek(int64(f.currentByte), 0)
	if err != nil {
		return 0, err
	}

	bytesWritten, err := f.file.Write(p)
	if err != nil {
		return 0, err
	}

	f.currentByte += bytesWritten

	fmt.Println("unlocked")
	fmt.Println("written: ", bytesWritten, " bytes")
	return bytesWritten, nil
}

func downloadFile(url string, chunkSize int, fileSize int, file *os.File) {
	mtx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for startByte := 0; startByte < fileSize; startByte += chunkSize {
		endByte := startByte + chunkSize - 1
		if endByte >= fileSize {
			endByte = fileSize - 1
		}

		wg.Add(1)
		go func(start, end int) {
			fmt.Println("start goroutine")
			defer wg.Done()

			reader, err := NewChunkDownloader(url, start, end)
			if err != nil {
				panic(err)
			}
			defer reader.Close()

			writer := NewFileChunkWriter(file, mtx, start)

			io.Copy(writer, reader)
			fmt.Println("end goroutine")
		}(startByte, endByte)
	}

	wg.Wait()
	fmt.Println("finished downloading")
}

func download(url string, destFile *os.File) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	fmt.Println("fileSize: ", fileSize)

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

func main() {
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
		log.Fatal(errors.New("invalid arguments"))
	}

	u, err := url.Parse(sourceUrl)
	if err != nil {
		log.Fatal(err)
	}

	fileName := filepath.Base(u.Path)

	var file *os.File
	defer file.Close()
	if destPath == "" {
		destPath, err = os.Getwd()
		file, err = os.Create(destPath + "/" + fileName)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := os.Stat(destPath)
		if err != nil {
			log.Fatal(err)
		}
		file, err = os.Create(destPath + "/" + fileName)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = download(sourceUrl, file)
	if err != nil {
		log.Fatal(err)
	}
}
