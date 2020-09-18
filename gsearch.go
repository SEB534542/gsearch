// file: list_posts.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type result struct {
	Title   string
	Snippet string
	Link    string
}

func main() {
	apiKey := "AIzaSyAkDRzH7lbbEJ2RRNDqezqo34dcPyRifUU"
	searchId := "08e2d4458e1f30b4c"
	q := "corona"
	d := 1
	customSearch(apiKey, searchId, q, d)
	results, err := customSearch(apiKey, searchId, q, d)
	if err != nil {
		log.Panic(err)
	}
	for i, v := range results {
		fmt.Printf("%v: %v | %v\n", i, v.Title, v.Link)
	}
}

func customSearch(apiKey, searchId, q string, d int) ([]*result, error) {
	// Get response from Google customsearch
	response, err := http.Get("https://www.googleapis.com/customsearch/v1?key=" + apiKey + "&cx=" + searchId + "&q=" + q + "&dateRestrict=d:" + fmt.Sprint(d))
	if err != nil {
		return nil, err
	}

	// Read response
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	m := map[string]interface{}{}
	err = json.Unmarshal(responseData, &m)
	if err != nil {
		return nil, err
	}

	// Get relevant data elements
	results := []*result{}
	for _, v := range m["items"].([]interface{}) {
		switch v := v.(type) {
		case map[string]interface{}:
			results = append(results, &result{
				Title:   v["title"].(string),
				Snippet: v["snippet"].(string),
				Link:    v["link"].(string),
			})
		default:
			return nil, fmt.Errorf("customSearch: Unknown data format in response")
		}
	}
	return results, nil
}
