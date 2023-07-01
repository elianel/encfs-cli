package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func extractKeyIvIdHost(url string) (string, string, int, string, error) {
	rgx := regexp.MustCompile(
		`https:\/\/([-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*))\/view#([a-zA-Z0-9=_]{32})-([a-zA-Z0-9=_]{16})-(\d+)`,
	)
	match := rgx.FindStringSubmatch(url)
	if match == nil {
		return "", "", 0, "", errors.New("invalid url. regex failure")
	}

	host := fmt.Sprintf("https://%s", match[1])
	key := match[3]
	iv := match[4]
	idString := match[5]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return "", "", 0, "", err
	}

	return key, iv, id, host, nil
}
func downloadFile(path string, downloadUrl string) error {
	key, iv, id, host, err := extractKeyIvIdHost(downloadUrl)
	if err != nil {
		return err
	}
	log.Printf("Extracted key:%s, iv:%s, id:%d, host:%s", key, iv, id, host)
	fileInfo, err := getFileInfo(id, host)
	if err != nil {
		return err
	}
	decryptedName, err := DecryptByteArray(HexStringToBytes(fileInfo.FileName), key, iv)
	if err != nil {
		return err
	}
	decryptedType, err := DecryptByteArray(HexStringToBytes(fileInfo.FileType), key, iv)
	if err != nil {
		return err
	}
	fileInfo.FileName = string(decryptedName)
	fileInfo.FileType = string(decryptedType)
	log.Printf(
		"FileInfo: id:%d, fileName:%s, fileType:%s, parts:%d",
		fileInfo.Id,
		fileInfo.FileName,
		fileInfo.FileType,
		fileInfo.Parts,
	)
	var file *os.File
	exists := true

	fileStat, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		exists = false
	}

	if exists {

		if fileStat.IsDir() {
			path = fmt.Sprintf("%s/%s", path, fileInfo.FileName)
		}
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
	} else {
		file, err = os.Create(path)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	var partBytes []byte
	for i := 0; i < fileInfo.Parts; i++ {
		partBytes, err = downloadPart(fileInfo.Id, i, host)
		if err != nil {
			return err
		}
		partBytes, err := DecryptByteArray(partBytes, key, iv)
		if err != nil {
			return err
		}
		_, err = file.Write(partBytes)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Download Successful. File saved to %s\n", path)
	return err
}
func getFileInfo(id int, host string) (*FileInfo, error) {
	url := fmt.Sprintf("%s/files/%d/info", host, id)
	log.Println("GetFileInfo:", url)

	resp, err := HttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("GetFileInfo::StatusCode != http.StatusOK")
	}

	var fileInfo FileInfo
	err = json.NewDecoder(resp.Body).Decode(&fileInfo)
	if err != nil {
	    return nil, err
	}
	return &fileInfo, nil
}
func downloadPart(id int, part int, host string) ([]byte, error) {
	url := fmt.Sprintf("%s/files/%d/%d", host, id, part)
	req, err := http.NewRequest("GET", url, nil)
	resp, err := requestWithRetries(&HttpClient, req, MaxRetries)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type FileInfo struct {
	Id       int    `json:"id"`
	FileName string `json:"fileName"`
	FileType string `json:"fileType"`
	Parts    int    `json:"parts"`
}
