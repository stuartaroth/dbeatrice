package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	configsDir := "configs"
	templatesDir := "templates"

	dirEntries, err := os.ReadDir(configsDir)
	if err != nil {
		log.Fatal(err)
	}

	var c Connector
	for _, entry := range dirEntries {
		filename := entry.Name()
		if !strings.Contains(filename, "example_") {
			bits := MustReadFile(fmt.Sprintf("%v/%v", configsDir, filename))
			c = ConnectorTypeFromBits(bits)
		}
	}

	if c == nil {
		log.Fatal("c is nil")
	}

	searchBits := MustReadFile(fmt.Sprintf("%v/search.html", templatesDir))
	searchTemplate := string(searchBits)

	searchResultsBits := MustReadFile(fmt.Sprintf("%v/search_results.html", templatesDir))
	searchResultsTemplate := string(searchResultsBits)

	http.HandleFunc("/search.html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(searchTemplate))
	})

	http.HandleFunc("/search_results.html", func(w http.ResponseWriter, r *http.Request) {
		var requestBody SearchResultsJson
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			writeError(w, err)
			return
		}

		qResults, err := c.Query(requestBody.Query)
		if err != nil {
			writeError(w, err)
			return
		}

		th := `<th scope="col">%v</th>`
		td := `<td>%v</td>`
		tr := `<tr>%v</tr>`

		ths := []string{}
		for _, h := range qResults.Headers {
			ths = append(ths, fmt.Sprintf(th, h.Name))
		}

		joinedThs := strings.Join(ths, "")

		trs := []string{}
		for _, row := range qResults.Rows {
			columns := []string{}
			for _, column := range row {
				columns = append(columns, fmt.Sprintf(td, column))
			}

			joinedColumns := strings.Join(columns, "")
			trs = append(trs, fmt.Sprintf(tr, joinedColumns))
		}

		joinedTrs := strings.Join(trs, "")

		localTemplate := searchResultsTemplate
		localTemplate = strings.Replace(localTemplate, "$THS", joinedThs, -1)
		localTemplate = strings.Replace(localTemplate, "$TRS", joinedTrs, -1)

		w.Write([]byte(localTemplate))
	})

	http.ListenAndServe(":8080", nil)
}

func writeError(w http.ResponseWriter, err error) {
	w.Write([]byte(fmt.Sprintf(`<div class="bg-danger">%v</div>`, err)))
}

type SearchResultsJson struct {
	Query string `json:"query"`
}

func MustReadFile(filename string) []byte {
	bits, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	return bits
}

func ConnectorTypeFromBits(bits []byte) Connector {
	var jsonConfig map[string]string
	err := json.Unmarshal(bits, &jsonConfig)
	if err != nil {
		log.Fatal(err)
	}

	typeValue, _ := jsonConfig["type"]

	if typeValue == "postgres" {
		connector, err := NewConnectorPostgres(jsonConfig)
		if err != nil {
			log.Fatal(err)
		}

		return connector
	} else {
		log.Fatal("provided type not supported:", typeValue)
		return nil
	}
}
