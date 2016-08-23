package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	ID    string
	Files []File
	Token string
	Type  string
}

// body represents the response from POEditor api.
type body struct {
	Response struct {
		Code    string
		Message string
		Status  string
	}
	Item string
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

// request will do a request to POEditor and get
// the download file url.
func request(token, typ, id, lang string) *body {
	form := url.Values{}

	if len(typ) == 0 {
		typ = "mo"
	}

	form.Add("api_token", token)
	form.Add("action", "export")
	form.Add("id", id)
	form.Add("type", typ)
	form.Add("language", lang)

	req, err := http.NewRequest("POST", "https://poeditor.com/api/", strings.NewReader(form.Encode()))
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

	for _, file := range config.Files {
		body := request(config.Token, config.Type, config.ID, file.Lang)

		if body == nil {
			continue
		}

		if body.Response.Code == "200" {
			downloadFromURL(file.Path, body.Item)
			fmt.Println(file.Path, "downloaded")
		} else {
			fmt.Println("Failed response from POEditor:", body.Response.Message, "-", file.Lang)
		}
	}
}
