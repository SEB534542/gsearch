// file: list_posts.go
package main

import (
	"encoding/csv"
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
	Date    string
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
var results = []*result{}
var port = ":8080"

const configFile = "config.json"
const exportFile = "output.csv"

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
	http.HandleFunc("/export", handlerExport)
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
		results = search()

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

func handlerExport(w http.ResponseWriter, req *http.Request) {
	// Transform output to [][]string
	lines := [][]string{}
	for _, v := range results {
		lines = append(lines, []string{v.Title, v.Date, v.Link})
	}
	// Write the file
	f, err := os.Create(exportFile)
	if err != nil {
		http.Error(w, "Error saving", http.StatusForbidden)
		return
	}
	defer f.Close()
	wr := csv.NewWriter(f)
	if err = wr.WriteAll(lines); err != nil {
		http.Error(w, "Error saving", http.StatusForbidden)
	}
	fmt.Fprintf(w, "Output saved as %s", exportFile)
	return
}

func search() []*result {
	// Page 1 (first ten results)
	results, totalResults, err := customSearch(1)
	if err != nil {
		log.Panic()
	}

	if totalResults > 10 {
		pages := totalResults/10 + 1
		// For each subsequent page, get the data
		for i := 2; i <= pages; i++ {
			r, _, err := customSearch(1)
			if err != nil {
				log.Panic()
			}
			results = append(results, r...)
		}
	}
	return results
}

func customSearch(page int) ([]*result, int, error) {
	// Get response from Google customsearch
	response, err := http.Get("https://www.googleapis.com/customsearch/v1?key=" + p.ApiKey + "&cx=" + p.SearchId + "&q=" + p.Query + "&dateRestrict=d" + fmt.Sprint(p.Days) + "&start=" + fmt.Sprint((page-1)*10+1))
	if err != nil {
		return nil, 0, err
	}

	// Read response
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal JSON
	m := map[string]interface{}{}
	err = json.Unmarshal(responseData, &m)
	if err != nil {
		return nil, 0, err
	}

	totalResults, err := strconv.Atoi(m["queries"].(map[string]interface{})["request"].([]interface{})[0].(map[string]interface{})["totalResults"].(string))
	if err != nil {
		log.Panic()
	}
	// Get relevant data elements
	results := []*result{}
	for _, v := range m["items"].([]interface{}) {
		switch v := v.(type) {
		case map[string]interface{}:
			date, ok := v["pagemap"].(map[string]interface{})["metatags"].([]interface{})[0].(map[string]interface{})["dc.date"].(string)
			if !ok {
				date, ok := v["pagemap"].(map[string]interface{})["metatags"].([]interface{})[0].(map[string]interface{})["article:published_time"].(string)
				if ok {
					date = date[:10]
				} else {
					date, ok := v["pagemap"].(map[string]interface{})["metatags"].([]interface{})[0].(map[string]interface{})["20060102"].(string)
					if ok {
						date = date[:4] + "-" + date[4:6] + "-" + date[6:]
					}
				}
			}
			results = append(results, &result{
				Title:   v["title"].(string),
				Snippet: v["snippet"].(string),
				Link:    v["link"].(string),
				Date:    date,
			})
		default:
			return nil, totalResults, fmt.Errorf("customSearch: Unknown data format in response")
		}
	}
	return results, totalResults, nil
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

	for _, v := range m["items"].([]interface{}) {
		switch v := v.(type) {
		case map[string]interface{}:
			date, ok := v["pagemap"].(map[string]interface{})["metatags"].([]interface{})[0].(map[string]interface{})["dc.date"].(string)
			if !ok {
				date, ok := v["pagemap"].(map[string]interface{})["metatags"].([]interface{})[0].(map[string]interface{})["article:published_time"].(string)
				if ok {
					date = date[:10]
				} else {
					date, ok := v["pagemap"].(map[string]interface{})["metatags"].([]interface{})[0].(map[string]interface{})["20060102"].(string)
					if ok {
						date = date[:4] + "-" + date[4:6] + "-" + date[6:]
					}
				}
			}
			fmt.Println(date)
		default:
			fmt.Println("No")
		}
	}
}
