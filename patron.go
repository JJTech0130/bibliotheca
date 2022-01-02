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

// Patron is a representation of a patron's session with a library
type Patron struct {
	Id      string
	Library *url.URL
	Client  http.Client
}

// NewPatron authenticates as the specified patron with the specified library
func NewPatron(id string, library *url.URL) (*Patron, error) {
	jar, _ := cookiejar.New(nil)

	client := http.Client{
		Jar:     jar,
		Timeout: time.Second * 30,
	}

	postBody, err := json.Marshal(map[string]string{"UserId": id})
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(library.String()+"/Patron/LoginPatron", "application/json", bytes.NewBuffer(postBody))
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
	} else if body["Success"] != true {
		return nil, errors.New(body["FailureReason"].(string))
	}

	p := Patron{Id: id, Library: library, Client: client}

	return &p, nil
}

type Action string // an action that can be preformed
const (
	Return Action = "Return"
	Borrow Action = "Borrow"
)

// itemAuthenticated a helper for getting raw data on an item, this version is authenticated so that AllowedPatronAction can work
func itemAuthenticated(id string, patron *Patron) (map[string]interface{}, error) {
	resp, err := patron.Client.Get(patron.Library.String() + "/Item/GetItem?id=" + id)
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

// AllowedAction returns the action(s) a patron can preform on a book.
func (p *Patron) AllowedAction(book *Book) (Action, error) {
	i, err := itemAuthenticated(book.Id, p)
	if err != nil {
		return "", err
	}

	log.Println(i)
	return Action(i["AllowedPatronAction"].(string)), nil
}

// borrowed helper for getting raw data on borrowed books
func borrowed(patron *Patron) ([]map[string]interface{}, error) {
	resp, err := patron.Client.Get(patron.Library.String() + "/Patron/Borrowed")
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

// Borrowed returns an array of the books currently borrowed by a patron.
func (p *Patron) Borrowed() ([]*Book, error) {
	b, err := borrowed(p)
	if err != nil {
		return nil, err
	}

	var books []*Book
	for _, s := range b {
		book := Book{
			Id:     s["Id"].(string),
			Title:  s["Title"].(string),
			Author: s["Authors"].(string),
			ISBN:   s["ISBN"].(string),
		}
		books = append(books, &book)
	}

	return books, nil
}

// Borrow borrows the specified book as the patron
func (p *Patron) Borrow(book *Book) error {
	a, err := p.AllowedAction(book)
	if err != nil {
		return err
	} else if a != Borrow {
		return errors.New("book cannot be borrowed")
	}

	postBody, _ := json.Marshal(map[string]string{"CatalogItemId": book.Id})

	resp, err := p.Client.Post(p.Library.String()+"/Item/Borrow", "application/json", bytes.NewBuffer(postBody))
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

// Return the specified book as the patron.
func (p *Patron) Return(book *Book) error {
	a, err := p.AllowedAction(book)
	if err != nil {
		return err
	} else if a != Return {
		return errors.New("book cannot be returned")
	}

	postBody, _ := json.Marshal(map[string]string{"CatalogItemId": book.Id})

	resp, err := p.Client.Post(p.Library.String()+"/Item/Return", "application/json", bytes.NewBuffer(postBody))
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

// obii helper to get the "OBII" of the specified book
func obii(book *Book, patron *Patron) (string, error) {
	b, err := borrowed(patron)
	if err != nil {
		return "", err
	}
	for _, s := range b {
		log.Println(s["Title"].(string) + ": " + s["Id"].(string) + ": " + s["Obii"].(string))
		if s["Id"] == book.Id {
			return s["Obii"].(string), nil
		}
	}
	return "", errors.New("book must be borrowed, can't generate OBII")
}

// Download an ASCM file for the specified book.
func (p *Patron) Download(book *Book) (string, error) {
	a, err := p.AllowedAction(book)
	if err != nil {
		return "", err
	} else if a != Return {
		return "", errors.New("book cannot be downloaded")
	}

	id, err := obii(book, p)
	if err != nil {
		return "", err
	}

	resp, err := p.Client.Get(p.Library.String() + "/Reader/OfflineReading?localEpub&id=" + id)
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
