package main

import (
	"errors"
	// "github.com/jlaffaye/ftp"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	// "time"
)

type Protocol int

const (
	UnknownProtocol Protocol = iota
	HTTP
	FTP
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

func checkProtocol(urlStr string) (Protocol, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return UnknownProtocol, err
	}

	scheme := strings.ToLower(u.Scheme)

	switch scheme {
	case "http", "https":
		return HTTP, nil
	case "ftp":
		return FTP, nil
	default:
		return UnknownProtocol, errors.New("unknown protocol")
	}
}

func checkArgs(download *Download) (Protocol, error) {
	_, err := os.Stat(download.path)
	if err != nil {
		return 0, errors.New("couldn't find given path")
	}

	prot, err := checkProtocol(download.url)
	if err != nil {
		return 0, errors.New("unknown protocol")
	}

	return prot, nil
}

func download(download *Download, protocol Protocol) error {
	switch protocol {
	case HTTP:
	case FTP:
	}
	return nil
}

func downloadHttp(download *Download) error {
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

// func parseFtpUrl(url string) error {
// 	return nil
// }
//
// func downloadFtp(download *Download) error {
// 	c, err := ftp.Dial(download.url, ftp.DialWithTimeout(5*time.Second))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	err = c.Login("anonymous", "anonymous")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	if err := c.Quit(); err != nil {
// 		log.Fatal(err)
// 	}
//
// 	return nil
// }

func main() {
	arg, err := getArgs()
	if err != nil {
		log.Println(err)
	}

	prot, err := checkArgs(&arg)
	if err != nil {
		log.Println(err)
	}

	err = download(&arg, prot)
	if err != nil {
		log.Default().Println(err)
	}
}
