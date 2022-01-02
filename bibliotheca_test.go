package bibliotheca

import (
	"log"
	"net/url"
	"testing"
)

func Test(t *testing.T) {
	baseURL, _ := url.Parse("https://ebook.yourcloudlibrary.com/uisvc/BethlehemDistrictLibraries")

	patron, err := NewPatron("11111", baseURL)
	if err != nil {
		t.Fatal(err)
	}

	b, err := NewBook("ammqdg9", baseURL)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Book: " + b.Title)
	log.Println("Author: " + b.Author)
	log.Println("ISBN: " + b.ISBN)

	allowed, err := patron.AllowedAction(b)
	if err != nil {
		t.Fatal(err)
	}

	if allowed == Return { // If we can return it, it must be borrowed
		log.Println("Downloading book...")
		ascm, err := patron.Download(b)
		if err != nil {
			t.Fatal(err)
		}
		log.Println(ascm)
	} else if allowed == Borrow {
		log.Println("Borrowing book...")
		err = patron.Borrow(b)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal("Book is in unknown state: " + allowed)
	}
}
