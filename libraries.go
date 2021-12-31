package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type Country string

// From bibCloud.js on yourcloudlibrary.com
// "3m.at", "Austria", "3m.au", "Australia", "3m.be", "Belgium", "3m.br", "Brazil", "3m.ca", "Canada", "3m.nz", "New Zealand", "3m.de", "Germany", "3m.il", "Israel", "3m.jp", "Japan", "3m.ro", "Romania", "3m.es", "Spain", "3m.za", "South Africa", "3m.sa", "Saudi Arabia", "3m.sg", "Singapore", "3m.ch", "Switzerland", "3m.ae", "UAE", "3m.gb", "United Kingdom", "3m.us", "United States"
const (
	Austria       Country = "3m.at"
	Australia     Country = "3m.au"
	Belgium       Country = "3m.be"
	Brazil        Country = "3m.br"
	Canada        Country = "3m.ca"
	NewZealand    Country = "3m.nz"
	Germany       Country = "3m.de"
	Israel        Country = "3m.il"
	Japan         Country = "3m.jp"
	Romania       Country = "3m.ro"
	Spain         Country = "3m.es"
	SouthAfrica   Country = "3m.za"
	SaudiArabia   Country = "3m.sa"
	Singapore     Country = "3m.sg"
	Switzerland   Country = "3m.ch"
	UAE           Country = "3m.ae"
	UnitedKingdom Country = "3m.gb"
	UnitedStates  Country = "3m.us"
)

func sendRPC(method string, params []string) (interface{}, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	jsonBody, err := json.Marshal(map[string]interface{}{"method": method, "params": params})
	if err != nil {
		return nil, err
	}

	resp, err := client.Get("https://service.yourcloudlibrary.com/json/rpc?json=" + url.QueryEscape(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	} else if body["error"] != nil {
		return nil, errors.New(body["error"].(map[string]interface{})["msg"].(string))
	}

	return body["result"], nil
}

func GetToken(country Country) (string, error) {
	resp, err := sendRPC("WSAuth.authenticateAnonymousUser", []string{string(country)})
	if err != nil {
		return "", err
	}

	token := resp.(map[string]interface{})["token"].(string)

	return token, nil
}

// State is an abbreviation for the state, e.g. PA or NY
type State string

func GetStates(country Country) ([]State, error) {
	resp, err := sendRPC("WSLibraryMgmt.getStates", []string{string(country)})
	if err != nil {
		return nil, err
	}

	stateArray := resp.([]interface{})
	var states []State
	for _, s := range stateArray {
		states = append(states, State(s.(map[string]interface{})["abbreviation"].(string)))
	}

	return states, nil
}

func containsState(s []State, e State) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Library name
type Library struct {
	Name string
	ID   string
}

func GetLibraries(country Country, state State) ([]Library, error) {
	states, err := GetStates(country)
	if err != nil {
		return nil, err
	} else if !containsState(states, state) {
		return nil, errors.New("state does not exist")
	}

	token, err := GetToken(country)
	if err != nil {
		return nil, err
	}

	resp, err := sendRPC("WSLibraryMgmt.getLibraryBranchesByState", []string{token, string(state)})
	if err != nil {
		return nil, err
	}

	libraryArray := resp.([]interface{})

	var libraries []Library
	for _, l := range libraryArray {
		libraries = append(libraries, Library{Name: l.(map[string]interface{})["name"].(string), ID: l.(map[string]interface{})["libraryID"].(string)})
	}

	return libraries, nil
}

func libraryByName(s []Library, e string) Library {
	for _, a := range s {
		if a.Name == e {
			return a
		}
	}

	return Library{}
}

func GetLibraryID(country Country, state State, name string) (string, error) {
	libraries, err := GetLibraries(country, state)
	if err != nil {
		return "", err
	}

	library := libraryByName(libraries, name)

	empty := Library{}
	if library == empty {
		return "", errors.New("library does not exist")
	}

	return library.ID, nil
}

func GetLibrary(country Country, id string) (map[string]interface{}, error) {
	token, err := GetToken(country)
	if err != nil {
		return nil, err
	}

	resp, err := sendRPC("WSLibraryMgmt.getLibraryByID", []string{token, id})
	if err != nil {
		return nil, err
	}

	return resp.(map[string]interface{}), nil
}

func GenerateURL(library map[string]interface{}) (*url.URL, error) {
	parsed, err := url.Parse("https://ebook.yourcloudlibrary.com/uisvc/" + library["urlName"].(string))
	if err != nil {
		return nil, err
	}
	return parsed, nil
}
