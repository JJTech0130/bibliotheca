package bibliotheca

import (
	"testing"
)

func TestLibraries(t *testing.T) {
	id, err := GetLibraryID(UnitedStates, "PA", "HELLERTOWN AREA LIBRARY")
	if err != nil {
		t.Error(err)
	}
	lib, err := GetLibrary(UnitedStates, id)
	if err != nil {
		t.Error(err)
	}
	baseURL, err := GenerateURL(lib)
	if err != nil {
		t.Error(err)
	}

	expected := "https://ebook.yourcloudlibrary.com/uisvc/BethlehemDistrictLibraries"
	if baseURL.String() != expected {
		t.Errorf("got " + baseURL.String() + " expected " + expected)
	}
}
