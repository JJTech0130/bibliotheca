package bibliotheca

import (
	"log"
	"net/url"
	"testing"
)

func Test(t *testing.T) {
	baseURL, _ := url.Parse("https://ebook.yourcloudlibrary.com/uisvc/BethlehemDistrictLibraries")

	bookId := "ammqdg9"

	session, err := Login("11111", baseURL)
	s := &session
	if err != nil {
		log.Fatal(err)
	}

	b, err := NewBook(bookId, s)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Book: " + b.Title)
	log.Println("Author: " + b.Author)
	log.Println("ISBN: " + b.ISBN)
	log.Println("Action: " + b.Action)

	if b.Action == Returnable {
		log.Println("Downloading book...")
		ascm, err := Download(bookId, s)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(ascm)
	} else if b.Action == Borrowable {
		log.Println("Borrowing book...")
		err = Borrow(bookId, s)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Book is in unknown state")
	}
}
