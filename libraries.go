package bibliotheca

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type Country string

// State is an abbreviation for the state, e.g. PA or NY
type State string

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

func tokenize(country Country) (string, error) {
	resp, err := sendRPC("WSAuth.authenticateAnonymousUser", []string{string(country)})
	if err != nil {
		return "", err
	}

	token := resp.(map[string]interface{})["token"].(string)

	return token, nil
}

func libraries(country Country, state State) (map[string]string, error) {
	token, err := tokenize(country)
	if err != nil {
		return nil, err
	}

	resp, err := sendRPC("WSLibraryMgmt.getLibraryBranchesByState", []string{token, string(state)})
	if err != nil {
		return nil, err
	}

	libraryArray := resp.([]interface{})

	libs := make(map[string]string)
	for _, l := range libraryArray {
		name := l.(map[string]interface{})["name"].(string)
		id := l.(map[string]interface{})["libraryID"].(string)
		libs[name] = id
	}

	return libs, nil
}

func libraryID(country Country, state State, name string) (string, error) {
	libs, err := libraries(country, state)
	if err != nil {
		return "", err
	}

	return libs[name], nil
}

func libraryDetails(country Country, state State, name string) (map[string]interface{}, error) {
	id, err := libraryID(country, state, name)
	if err != nil {
		return nil, err
	}

	token, err := tokenize(country)
	if err != nil {
		return nil, err
	}

	resp, err := sendRPC("WSLibraryMgmt.getLibraryByID", []string{token, id})
	if err != nil {
		return nil, err
	}

	return resp.(map[string]interface{}), nil
}

func LibraryURL(country Country, state State, name string) (*url.URL, error) {
	lib, err := libraryDetails(country, state, name)
	if err != nil {
		return nil, err
	}

	u := url.URL{Scheme: "https", Host: "ebook.yourcloudlibrary.com", Path: "uisvc"}
	u.Path = u.Path + "/" + lib["urlName"].(string)

	return &u, nil
}
