package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

var commands = []string{
	"encfs-cli [flags] [upload|u] </path/to/file> <apiKey>",
	"encfs-cli [flags] [download|d] </path/to/download/to> <downloadUrl>",
}

func PrintHelp() {
	fmt.Println("Invalid Usage")
	for _, c := range commands {
		fmt.Println(c)
	}
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func main() {
	host := flag.String("h", DefaultHost, "Custom host")
	verboseLogging := flag.Bool("v", false, "Verbose logging.")
	flag.Parse()
	if flag.NArg() < 3 {
		PrintHelp()
		return
	}
	if !*verboseLogging {
		nullDev, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(nullDev)
	}
	action := flag.Arg(0)

	switch action {
	case "upload", "u":
		filePath := flag.Arg(1)
		apiKey := flag.Arg(2)
		err := handleUpload(apiKey, *host, filePath)
		if err != nil {
			log.Fatal(err)
		}
	case "download", "d":
		dlPath := flag.Arg(1)
		downloadUrl := flag.Arg(2)
		if downloadUrl == "" {
			PrintHelp()
		}
		err := handleDownload(dlPath, downloadUrl)
		if err != nil {
			log.Fatal(err)
		}
	default:
		PrintHelp()
	}

}
func handleDownload(filePath string, downloadUrl string) error {
	if filePath == "" {
		return errors.New("file path not set")
	}
	if downloadUrl == "" {
		return errors.New("download url not set")
	}
	return downloadFile(filePath, downloadUrl)
}
func handleUpload(apiKey string, host string, filePath string) error {
	if apiKey == "" {
		return errors.New("API key not set")
	}
	if filePath == "" {
		return errors.New("file path not set")
	}
	var h string
	if host == "" {
		h = DefaultHost
	} else {
		h = host
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return errors.New("can't upload a dir")
	}
	return uploadFile(filePath, "https://"+h, apiKey)
}
