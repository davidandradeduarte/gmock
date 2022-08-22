package gmock // nolint:golint

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// addStubHandler is the handler for the /httpmock/add endpoint.
// It expects a POST request with a JSON or YAML body containing a StubRequest.
// The stub is added to the server's stubs map.
// If the stub is invalid, it returns a 400 Bad Request.
// If the stub is valid, it returns a 201 Created.
func (s *Server) addStubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Error().Msgf("method %s not allowed", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Msgf("error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stub, err := getStubFromBytes(body)
	if err != nil {
		log.Error().Msgf("failed to parse stub request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if errs := s.addStub(stub); len(errs) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// listStubsHandler is the handler for the /httpmock/list endpoint.
// It expects a GET request.
// It returns a JSON array with the existing stubs.
// If no stubs are found, it returns an empty array.
func (s *Server) listStubsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Error().Msgf("method %s not allowed", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	values := make([]*Stub, 0, len(s.stubs))
	for _, v := range s.stubs {
		values = append(values, v)
	}

	body, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		log.Error().Msgf("error marshaling stubs: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(body); err != nil {
		log.Error().Msgf("error listing stubs: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// genericStubHandler is the handler for any stub endpoint.
// It returns the stub response if the current request is an existing stub request.
// If the request is not a stub request, it returns a 404 Not Found.
func (s *Server) genericStubHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Msgf("error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// compact JSON body before hashing
	var compactedBody string
	if body != nil {
		buff := new(bytes.Buffer)
		if err := json.Compact(buff, body); err != nil {
			log.Error().Msgf("error compacting body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		compactedBody = buff.String()
	}

	h := getHash(StubRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Body:   compactedBody,
	})
	if stub, ok := s.stubs[h]; ok {
		for k, v := range stub.Response.Headers {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		if _, ok := w.Header()["Content-Type"]; !ok {
			w.Header().Set("Content-Type", defaultContentType)
		}
		log.Info().Msgf("stub found: %s", stub.Request.String())
		body, err := json.Marshal(stub.Response.Body)
		if err != nil {
			log.Error().Msgf("error marshaling stub body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(body); err != nil {
			log.Error().Msgf("error writing stub body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if stub.Response.StatusCode != http.StatusOK {
			w.WriteHeader(stub.Response.StatusCode)
		}
		return
	}
	log.Error().Msgf("no stub found for request: %s %s %s", r.Method, r.URL.String(), compactedBody)
	w.WriteHeader(http.StatusNotFound)
}
