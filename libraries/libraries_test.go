package libraries

import (
	"testing"
)

func Test(t *testing.T) {
	baseURL, err := LibraryURL(UnitedStates, "PA", "BETHLEHEM AREA PUBLIC LIBRARY")
	if err != nil {
		t.Fatal(err)
	}

	if baseURL.String() != "https://ebook.yourcloudlibrary.com/uisvc/BethlehemDistrictLibraries" {
		t.Errorf("unexpected URL " + baseURL.String())
	}
}
