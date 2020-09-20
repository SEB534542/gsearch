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

type output struct {
	SearchInformation struct {
		TotalResults string `json:"totalResults"`
	} `json:"searchInformation"`
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
	parameters
}

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
var o = &output{}
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
		seb.SaveToJSON(o.parameters, configFile)
	} else {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatalf("%s is corrupt. Please delete the file (%v)", configFile, err)
		}
		err = json.Unmarshal(data, o)
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
	if req.Method == http.MethodPost {
		var err error
		o.ApiKey = req.PostFormValue("ApiKey")
		o.SearchId = req.PostFormValue("SearchId")
		o.Query = req.PostFormValue("Query")
		o.Days, err = strconv.Atoi(req.PostFormValue("Days"))
		if err != nil {
			http.Error(w, "Please enter a number of Days", http.StatusForbidden)
			return
		}
		seb.SaveToJSON(o.parameters, configFile)
		err = o.search()
		if err != nil {
			msg := "Error: " + fmt.Sprint(err)
			http.Error(w, msg, http.StatusForbidden)
			return
		}
	}
	err := tpl.ExecuteTemplate(w, "index.gohtml", o)
	if err != nil {
		log.Panic(err)
	}
}

func handlerExport(w http.ResponseWriter, req *http.Request) {
	err := o.export(exportFile)
	if err != nil {
		msg := "Error saving:" + fmt.Sprint(err)
		http.Error(w, msg, http.StatusBadRequest)
	}
	fmt.Fprintf(w, "Output saved as %s", exportFile)
	return
}

func (o *output) export(fname string) error {
	// Transform output to [][]string
	lines := [][]string{}
	for _, v := range o.Items {
		lines = append(lines, []string{v.Title, v.Link})
	}
	// Write the file
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	wr := csv.NewWriter(f)
	if err = wr.WriteAll(lines); err != nil {
		return err
	}
	return nil
}

func (o *output) search() error {
	// Empty items and SearchInformation
	o.Items = o.Items[:0]
	o.SearchInformation.TotalResults = "0"

	// Get Page 1 (first ten results) + number of totalResults
	err := o.customSearch(1)
	if err != nil {
		return err
	}
	r, err := strconv.Atoi(o.SearchInformation.TotalResults)
	if err != nil {
		return err
	}
	if r > 10 {
		pages := r/10 + 1
		// For each subsequent page, get the data
		for p := 2; p <= pages; p++ {
			o.customSearch(p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *output) customSearch(page int) error {
	// Get response from Google customsearch
	url := "https://www.googleapis.com/customsearch/v1?key=" + o.ApiKey + "&cx=" + o.SearchId + "&q=" + o.Query + "&dateRestrict=d" + fmt.Sprint(o.Days) + "&start=" + fmt.Sprint((page-1)*10+1)
	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	// Read response
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	if o.Items == nil {
		err = json.Unmarshal(responseData, o)
		if err != nil {
			return err
		}
		return nil
	}

	tempOutput := output{}
	err = json.Unmarshal(responseData, &tempOutput)
	if err != nil {
		return err
	}

	o.Items = append(o.Items, tempOutput.Items...)
	return nil
}

func temp() {
	// Read response
	responseData, err := ioutil.ReadFile("output.json")
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal JSON
	o := &output{}
	err = json.Unmarshal(responseData, o)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Total Results:", o.SearchInformation.TotalResults)
	for _, v := range o.Items {
		fmt.Println(v.Title, v.Link)
	}
}
