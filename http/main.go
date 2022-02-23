// TODO:
// - support multiple mocks for same request
// - support multiple mocks from the same file
// - support wildcards in request path
// - support methos with contains/equals/starts-with/ends-with/etc
// - support yaml/json
// - support request/response headers
// - support request/response cookies
// nice to have: make a UI to configure the mock server and list existing mocks

package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type hash string

type Mocks map[hash]Mock

type Mock struct {
	Request  MockRequest  `json:"request"`
	Response MockResponse `json:"response"`
}

type MockRequest struct {
	Method string     `json:"method"`
	Path   string     `json:"path"`
	Query  url.Values `json:"query"`
}

type MockResponse struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
}

const (
	defaultStubsLocation = "mocks"
)

var (
	mocks = Mocks{}
)

func init() {
	stubs := flag.String("l", defaultStubsLocation, "stubs location")
	flag.Parse()

	files, err := ioutil.ReadDir(*stubs)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fileBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", *stubs, file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		var mock Mock
		err = json.Unmarshal(fileBytes, &mock)
		if err != nil {
			log.Fatal(err)
		}

		mocks[getHash(mock.Request)] = mock
	}
}

func main() {
	http.HandleFunc("/gmock/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("[ERROR] Method %s not allowed", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("[ERROR] Error reading body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var mock Mock
		err = json.Unmarshal(body, &mock)
		if err != nil {
			log.Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// todo: validate mock (e.g valid status code, body, etc)

		h := getHash(mock.Request)
		if _, ok := mocks[h]; ok {
			log.Printf("[ERROR] Mock already exists")
			w.WriteHeader(http.StatusConflict)
			return
		}

		mocks[h] = mock
		log.Printf("[INFO] Added mock for %s", mock.Request.Path)
		w.WriteHeader(http.StatusCreated)
	})
	http.HandleFunc("/gmock/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("[ERROR] Method %s not allowed", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		values := make([]Mock, 0, len(mocks))
		for _, v := range mocks {
			values = append(values, v)
		}

		body, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			log.Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(body)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h := getHash(MockRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
		})
		if mock, ok := mocks[h]; ok {
			log.Printf("[INFO] Responding with mock for %s", mock.Request.Path)
			w.WriteHeader(mock.Response.StatusCode)
			body, err := json.Marshal(mock.Response.Body)
			if err != nil {
				log.Printf("[ERROR] %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(body)
			return
		}
		log.Printf("[INFO] No mock found for %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func getHash(r MockRequest) hash {
	return hash(AsSha256(r))
}

func AsSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
