package bibliotheca

import "log"

type Action string

const (
	Returnable Action = "Return"
	Borrowable Action = "Borrow"
)

type Book struct {
	Title  string
	Author string
	ISBN   string
	Action
}

func NewBook(id string, session *Session) (Book, error) {
	i, err := item(id, session)
	if err != nil {
		return Book{}, err
	}

	log.Println(i)
	return Book{
		Title:  i["Title"].(string),
		Author: i["Authors"].(string),
		ISBN:   i["ISBN"].(string),
		Action: Action(i["AllowedPatronAction"].(string)),
	}, nil
}
