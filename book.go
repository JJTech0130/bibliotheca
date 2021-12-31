package bibliotheca

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
)

// State describes the state of the book, in terms of the actions you can preform on it
type State string

const (
	Returnable State = "Return"
	Borrowable State = "Borrow"

	Borrowed  State = "Return" // Alternate wording for Returnable
	Available State = "Borrow" // Alternate wording for Borrowable
)

type Book struct {
	Id     string
	Title  string
	Author string
	ISBN   string
	State
}

func NewBook(id string, session *Session) (Book, error) {
	i, err := item(id, session)
	if err != nil {
		return Book{}, err
	}

	log.Println(i)
	return Book{
		Id:     id,
		Title:  i["Title"].(string),
		Author: i["Authors"].(string),
		ISBN:   i["ISBN"].(string),
		State:  State(i["AllowedPatronAction"].(string)),
	}, nil
}

// Borrow borrows this book from the library. The book must be borrowable.
func (b *Book) Borrow(session *Session) error {
	if b.State != Borrowable {
		return errors.New("book is not borrowable")
	}

	postBody, _ := json.Marshal(map[string]string{"CatalogItemId": b.Id})

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

// Return returns this book to the library. The book must be returnable.
func (b *Book) Return(session *Session) error {
	if b.State != Returnable {
		return errors.New("book is not returnable")
	}

	postBody, _ := json.Marshal(map[string]string{"CatalogItemId": b.Id})

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

// Download downloads an ASCM file. The book must be borrowed.
func (b *Book) Download(session *Session) (string, error) {
	if b.State != Borrowed {
		return "", errors.New("book is not borrowed")
	}

	id, err := obii(b.Id, session)
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
