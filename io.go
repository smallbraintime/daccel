package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func NewChunkDownloader(url string, startByte int, endByte int, client *http.Client) (*ChunkDownloader, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &ChunkDownloader{resp.Body}, nil
}

type ChunkDownloader struct {
	reader io.ReadCloser
}

func (c *ChunkDownloader) Read(p []byte) (n int, err error) {
	bytesRead, err := c.reader.Read(p)
	if err != nil {
		return 0, err
	}

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
	file   *os.File
	mtx    *sync.Mutex
	offset int
}

func (f *FileChunkWriter) Write(p []byte) (n int, err error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	_, err = f.file.Seek(int64(f.offset), 0)
	if err != nil {
		return 0, err
	}

	bytesWritten, err := f.file.Write(p)
	if err != nil {
		return 0, err
	}

	f.offset += bytesWritten

	return bytesWritten, nil
}
