package bibliotheca

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Book struct {
	Id     string
	Title  string
	Author string
	ISBN   string
}

// item a helper for getting raw data on an item
func item(id string, library *url.URL) (map[string]interface{}, error) {

	client := http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Get(library.String() + "/Item/GetItem?id=" + id)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	} else if body["ErrorCode"] != nil {
		return nil, errors.New(body["ErrorMessage"].(string))
	}

	return body, nil
}

func NewBook(id string, library *url.URL) (*Book, error) {
	i, err := item(id, library)
	if err != nil {
		return nil, err
	}

	log.Println(i)
	return &Book{
		Id:     id,
		Title:  i["Title"].(string),
		Author: i["Authors"].(string),
		ISBN:   i["ISBN"].(string),
	}, nil
}
