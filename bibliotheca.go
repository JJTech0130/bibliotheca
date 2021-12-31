package bibliotheca

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

// Login to the Bibliotheca cloudLibrary at the specified URL, with the specified userId.
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

// item gets a raw map of information on the specified item
func item(id string, session *Session) (map[string]interface{}, error) {
	resp, err := session.Client.Get(session.URL.String() + "/Item/GetItem?id=" + id)
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

// Borrow borrows the specified book from the library. The book must be available.
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

// Return returns the specified book to the library. The book must be borrowed.
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

// Session holds the URL to the library, as well as a Client containing the session cookie
type Session struct {
	URL    url.URL
	Client http.Client
}

// borrowed gets a raw map of information on borrowed books
func borrowed(session *Session) ([]map[string]interface{}, error) {
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

// obii gets the "OBII" for the specified book, necessary for downloading.
// the book must be borrowed in order for this to work.
func obii(id string, session *Session) (string, error) {
	borrowed, err := borrowed(session)
	if err != nil {
		return "", err
	}
	for _, s := range borrowed {
		log.Println(s["Title"].(string) + ": " + s["Id"].(string) + ": " + s["Obii"].(string))
		if s["Id"] == id {
			return s["Obii"].(string), nil
		}
	}
	return "", errors.New("book must be borrowed, can't generate OBII")
}

// Download downloads an ASCM file for the specified book. It must be borrowed for this to work.
func Download(id string, session *Session) (string, error) {
	id, err := obii(id, session)
	if err != nil {
		return "", err
	}

	resp, err := session.Client.Get(session.URL.String() + "/Reader/OfflineReading?localEpub&id=" + id)
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
