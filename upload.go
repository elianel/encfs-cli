package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func uploadFile(filePath string, host string, apiKey string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileName := filepath.Base(filePath)
	fileSize := fileInfo.Size()
	parts := int(math.Ceil(float64(fileSize) / float64(PartSize)))
	fileType := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(fileType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
    // Trim the MIME type to keep only the first part before the semicolon, example: text/plain; charset=utf-8
	if index := strings.Index(mimeType, ";"); index != -1 {
		mimeType = mimeType[:index]
	}
	fileBytes := make([]byte, fileSize)
	_, err = file.Read(fileBytes)
	if err != nil {
		return err
	}
	log.Println("Parts:", parts)
	log.Println("Mime type:", mimeType)

	encKey := RandomString(32)
	iv := RandomString(16)

	encName, err := EncryptByteArray([]byte(fileName), encKey, iv)
	if err != nil {
		return err
	}
	encNameStr := hex.EncodeToString(encName)
	encMime, err := EncryptByteArray([]byte(mimeType), encKey, iv)
	if err != nil {
		return err
	}
	encMimeStr := hex.EncodeToString(encMime)

	fileCreateResp, err := createFile(apiKey, host, encNameStr, encMimeStr, strconv.Itoa(parts))
	if err != nil {
		return err
	}
	fileID := fileCreateResp.ID

	// Upload file parts
	for i := 0; i < parts; i++ {
		startLocation := i * PartSize
		endLocation := startLocation + PartSize
		if endLocation > len(fileBytes) {
			endLocation = len(fileBytes)
		}
		data := fileBytes[startLocation:endLocation]
		encData, err := EncryptByteArray(data, encKey, iv)
		if err != nil {
			return err
		}
		err = uploadPart(apiKey, host, fileID, i, &encData)
		if err != nil {
			log.Println("Failed to upload file part:", err)
			return err

		}
		fileSize -= int64(PartSize)
	}
	fmt.Printf("%s/view#%s-%s-%d\n", host, encKey, iv, fileID) //$"{config.Host}/view#{encKey}-{iv}-{fileId}
	return nil
}

type FileCreateResp struct {
	ID int64 `json:"id"`
}

func uploadPart(apiKey string, host string, fileID int64, part int, data *[]byte) error {
	partURL := fmt.Sprintf("%s/files/%d/%d", host, fileID, part)
	uploadPartReq, err := http.NewRequest("POST", partURL, bytes.NewReader(*data))
	if err != nil {
		return err
	}
	uploadPartReq.Header.Set("key", apiKey)
	_, err = requestWithRetries(&HttpClient, uploadPartReq, MaxRetries)
	return err
}
func createFile(apiKey string, host string, fileName string, fileType string, parts string) (*FileCreateResp, error) {
	createFileURL := fmt.Sprintf("%s/files/create", host)
	createFileReq, err := http.NewRequest("GET", createFileURL, nil)
	if err != nil {
		return nil, err
	}
	createFileReq.Header.Set("key", apiKey)
	createFileReq.Header.Set("fileName", fileName)
	createFileReq.Header.Set("fileType", fileType)
	createFileReq.Header.Set("parts", parts)

	fileCreateResponse, err := HttpClient.Do(createFileReq)
	if err != nil {
		return nil, err
	}
	defer fileCreateResponse.Body.Close()

	fileCreateRespData, err := ioutil.ReadAll(fileCreateResponse.Body)
	if err != nil {
		return nil, err
	}

	var fileCreateResp FileCreateResp
	err = json.Unmarshal(fileCreateRespData, &fileCreateResp)
	return &fileCreateResp, err
}
