// file: list_posts.go
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/SEB534542/seb"
	"github.com/pkg/browser"
)

type result struct {
	Title   string
	Snippet string
	Link    string
}

type parameters struct {
	ApiKey   string
	SearchId string
	Query    string
	Days     int
}

var tpl *template.Template
var p = parameters{}
var port = ":8080"

const configFile = "config.json"

func init() {
	// Loading gohtml templates
	var err error
	tpl, err = template.ParseFiles("index.gohtml")
	if err != nil {
		log.Fatal("Error loading html template")
	}
}

func main() {
	log.Println("--------Start of program--------")

	// Load parameters
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// File does not exist, creating blank
		seb.SaveToJSON(parameters{"", "", "", 0}, configFile)
	} else {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatalf("%s is corrupt. Please delete the file (%v)", configFile, err)
		}
		err = json.Unmarshal(data, &p)
		if err != nil {
			log.Fatalf("%s is corrupt. Please delete the file (%v)", configFile, err)
		}
	}

	err := browser.OpenURL("http://localhost" + port)
	if err != nil {
		log.Printf("Unable to open browser. Please visit: http://localhost%v in your browser", port)
	}

	log.Println("Launching website...")
	http.HandleFunc("/", handlerMain)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	log.Fatal(http.ListenAndServe(port, nil))
}

func handlerMain(w http.ResponseWriter, req *http.Request) {
	data := struct {
		Results []*result
		parameters
	}{
		parameters: p,
	}
	if req.Method == http.MethodPost {
		var err error
		p.ApiKey = req.PostFormValue("ApiKey")
		p.SearchId = req.PostFormValue("SearchId")
		p.Query = req.PostFormValue("Query")
		p.Days, err = strconv.Atoi(req.PostFormValue("Days"))
		if err != nil {
			http.Error(w, "Please enter a number of Days", http.StatusForbidden)
			return
		}
		seb.SaveToJSON(p, configFile)
		results, err := customSearch(p.ApiKey, p.SearchId, p.Query, p.Days)
		if err != nil {
			http.Error(w, "Error processing search, please check connection, API key and Search ID", http.StatusForbidden)
			return
		}
		data.Results = results
		data.parameters = p
	}
	err := tpl.ExecuteTemplate(w, "index.gohtml", data)
	if err != nil {
		log.Panic(err)
	}
}

func customSearch(apiKey, searchId, q string, d int) ([]*result, error) {
	// Get response from Google customsearch
	response, err := http.Get("https://www.googleapis.com/customsearch/v1?key=" + apiKey + "&cx=" + searchId + "&q=" + q + "&dateRestrict=d" + fmt.Sprint(d))
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

func temp() {
	// Read response
	responseData, err := ioutil.ReadFile("output.json")
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal JSON
	m := map[string]interface{}{}
	err = json.Unmarshal(responseData, &m)
	if err != nil {
		fmt.Println(err)
	}

	totalResults := m["queries"].(map[string]interface{})["request"].([]interface{})[0].(map[string]interface{})["totalResults"]

	fmt.Println(rx)

	result := m["queries"].(map[string]interface{})
	fmt.Printf("%T\n", result)
	r2 := result["request"].([]interface{})
	fmt.Printf("%T\n", r2[0])
	r3 := r2[0]
	fmt.Println(r3.(map[string]interface{})["totalResults"])

}
