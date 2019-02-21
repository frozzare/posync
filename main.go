package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	configFlag = flag.String("config", "", "Path to config file")
)

// File represents a config file struct.
type File struct {
	Lang string
	Path string
}

// Config represents a config struct.
type Config struct {
	Download bool
	ID       string
	Files    []File
	Path     string
	Token    string
	Type     string
	Upload   bool
}

// body represents the response from POEditor api.
type body struct {
	Response struct {
		Code    string
		Message string
		Status  string
	}
	Result struct {
		URL string
	}
}

// getConfig will return config instance.
func getConfig() Config {
	path := "config.json"

	if len(*configFlag) > 0 {
		path = *configFlag
	}

	file, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Printf("File error: %v\n", err)
		return Config{}
	}

	var config Config
	json.Unmarshal(file, &config)
	return config
}

// uploadRequest will do a request to POEditor with pot file.
func uploadRequest(token, id, path string) {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("Error while reading file", err)
		return
	}
	defer file.Close()

	form := &bytes.Buffer{}
	writer := multipart.NewWriter(form)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		fmt.Println("Error while doing writer", err)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error while doing copy", err)
		return
	}

	writer.WriteField("api_token", token)
	writer.WriteField("id", id)
	writer.WriteField("updating", "terms")

	err = writer.Close()
	if err != nil {
		fmt.Println("Error while doing writer", err)
		return
	}

	req, err := http.NewRequest("POST", "https://api.poeditor.com/v2/projects/upload", form)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	if err != nil {
		fmt.Println("Error while creating request", err)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error while doing request", err)
		return
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var b *body
	err = decoder.Decode(&b)

	if err != nil {
		fmt.Println("Error while decoding", err)
		return
	}

	fmt.Println("Upload response:", b.Response.Status, "-", b.Response.Message)
}

// downloadRequest will do a request to POEditor and get
// the download file url.
func downloadRequest(token, id, lang, typ string) *body {
	form := url.Values{}

	form.Add("api_token", token)
	form.Add("id", id)
	form.Add("type", typ)
	form.Add("language", lang)

	req, err := http.NewRequest("POST", "https://api.poeditor.com/v2/projects/export", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		fmt.Println("Error while creating request", err)
		return nil
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error while doing request", err)
		return nil
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var b *body
	err = decoder.Decode(&b)

	if err != nil {
		fmt.Println("Error while decoding", err)
		return nil
	}

	return b
}

// downloadFromURL will download a file from url to a file.
func downloadFromURL(path, url string) {
	output, err := os.Create(path)
	if err != nil {
		fmt.Println("Error while creating", path, "-", err)
		return
	}
	defer output.Close()

	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer res.Body.Close()

	_, err = io.Copy(output, res.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
}

func main() {
	flag.Parse()

	config := getConfig()

	if config.Upload {
		uploadRequest(config.Token, config.ID, config.Path)
	}

	if config.Download {
		for _, file := range config.Files {
			body := downloadRequest(config.Token, config.ID, file.Lang, config.Type)

			if body == nil {
				continue
			}

			if body.Response.Code == "200" {
				downloadFromURL(file.Path, body.Result.URL)
				fmt.Println("Download response:", file.Path, "downloaded")
			} else {
				fmt.Println("Failed response from POEditor:", body.Response.Message, "-", file.Lang)
			}
		}
	}
}
