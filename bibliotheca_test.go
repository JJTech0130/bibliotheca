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

	item, err := GetItem(bookId, s)
	if err != nil {
		log.Fatal(err)
	}

	borrowed, err := Borrowed(s)
	log.Println(borrowed)
	test, _ := Download(bookId, s)
	log.Println(test)

	//log.Println(item)
	log.Println("Book: " + item["Title"].(string))
	log.Println("ISBN: " + item["ISBN"].(string))
	action := item["AllowedPatronAction"].(string)
	log.Println("Action: " + action)

	if action == "Return" {
		log.Println("Returning book...")
		/*err = Return(bookId, s)
		if err != nil {
			log.Fatal(err)
		}*/
	} else if action == "Borrow" {
		log.Println("Borrowing book...")
		err = Borrow(bookId, s)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Could not borrow or return book")
	}
}
