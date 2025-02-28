package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Download struct {
	url, path string
}

func download(download *Download) error {
	resp, err := http.Get(download.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	index := strings.LastIndex(download.url, "/")
	if index == -1 {
		return errors.New("invalid url")
	}

	var file *os.File
	defer file.Close()
	if download.path != "" && download.path != "." {
		file, err = os.Create(download.path + download.url[index:])
		if err != nil {
			return err
		}
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}

		file, err = os.Create(currentDir + download.url[index:])
		if err != nil {
			return err
		}
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

func main() {
	args := os.Args
	if len(args) > 3 || len(args) == 2 {
		log.Fatal(errors.New("Invalid Arguments"))
	}

	if len(args) < 2 {
		fmt.Printf("daccel 0.1\nUsage: daccel <url> <dest>")
		return
	}

	arg := Download{
		args[1],
		args[2],
	}

	_, err := os.Stat(arg.path)
	if err != nil {
		log.Fatal(errors.New("couldn't find given path"))
	}

	err = download(&arg)
	if err != nil {
		log.Default().Println(err)
	}
}
