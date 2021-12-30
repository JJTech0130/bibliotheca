package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Login to the Bibliotheca cloudLibrary at the specified URL, with the specified userId
// Returns a Session for use with later requests
func Login(userId string, baseURL *url.URL) (Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Failed to create empty cookie jar!")
	}

	client := http.Client{
		Jar:     jar,
		Timeout: time.Second * 10,
	}

	postBody, _ := json.Marshal(map[string]string{"UserId": userId})

	resp, err := client.Post(baseURL.String()+"/Patron/LoginPatron", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return Session{}, err
	}

	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return Session{}, err
	} else if body["ErrorCode"] != nil {
		return Session{}, errors.New(body["ErrorMessage"].(string))
	} else if body["Success"] != true {
		return Session{}, errors.New(body["FailureReason"].(string))
	}

	return Session{*baseURL, client}, nil
}

func GetItem(id string, session *Session) (map[string]interface{}, error) {
	resp, err := session.Client.Get(session.URL.String() + "/Item/GetItem?id=" + id)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// This is terribly inefficient, should come up with a better way to handle errors
	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	} else if body["ErrorCode"] != nil {
		return nil, errors.New(body["ErrorMessage"].(string))
	}

	return body, nil
}

func Borrow(id string, session *Session) error {
	postBody, _ := json.Marshal(map[string]string{"CatalogItemId": id})

	resp, err := session.Client.Post(session.URL.String()+"/Item/Borrow", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return err
	} else if body["ErrorCode"] != nil {
		return errors.New(body["ErrorMessage"].(string))
	} else if body["Result"] != true {
		return errors.New(body["Message"].(string))
	}

	return nil
}

func Return(id string, session *Session) error {
	postBody, _ := json.Marshal(map[string]string{"CatalogItemId": id})

	resp, err := session.Client.Post(session.URL.String()+"/Item/Return", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return err
	} else if body["ErrorCode"] != nil {
		return errors.New(body["ErrorMessage"].(string))
	} else if body["Result"] != true {
		return errors.New(body["Message"].(string))
	}

	return nil
}

type Session struct {
	URL    url.URL
	Client http.Client
}

func Borrowed(session *Session) ([]map[string]interface{}, error) {
	resp, err := session.Client.Get(session.URL.String() + "/Patron/Borrowed")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var body []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func Obii(id string, session *Session) (string, error) {
	borrowed, err := Borrowed(session)
	if err != nil {
		return "", err
	}
	for _, s := range borrowed {
		log.Println(s["Title"].(string) + ": " + s["Id"].(string) + ": " + s["Obii"].(string))
		if s["Id"] == id {
			return s["Obii"].(string), nil
		}
	}
	return "", errors.New("book not borrowed")
}

func Download(id string, session *Session) (string, error) {
	obii, err := Obii(id, session)
	if err != nil {
		return "", err
	}

	resp, err := session.Client.Get(session.URL.String() + "/Reader/OfflineReading?localEpub&id=" + obii)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)

	return sb, nil
}
