package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

type Download struct {
	url, path string
}

func getArgs() (Download, error) {
	args := os.Args
	if len(args) > 3 || len(args) < 2 {
		return Download{}, errors.New("Invalid Arguments")
	}
	return Download{
		args[1],
		args[2],
	}, nil
}

func download(download *Download) error {
	resp, err := http.Get(download.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(download.path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func main() {
	arg, err := getArgs()
	if err != nil {
		log.Println(err)
	}

	_, err = os.Stat(arg.path)
	if err != nil {
		log.Println(errors.New("couldn't find given path"))
	}

	err = download(&arg)
	if err != nil {
		log.Default().Println(err)
	}
}
