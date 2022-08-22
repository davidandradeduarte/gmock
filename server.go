package gmock // nolint:golint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	defaultStubsDir    = "stubs"            // the default location for stubs
	defaultContentType = "application/json" // the default content type for stubs
	defaultPort        = 8080               // the default port for the server
)

// Config is used to configure the server.
type Config struct {
	Port     int
	StubsDir string
	Stubs    []*Stub
}

// Server is the HTTP mock server.
type Server struct {
	stubs stubHashMap
	dir   string
	port  int
	srv   *http.Server
}

// NewServer creates a new server.
func NewServer() *Server {
	s := &Server{
		stubs: make(stubHashMap),
		dir:   defaultStubsDir,
		port:  defaultPort,
	}
	return s
}

// NewServerWithConfig creates a new server with the given config.
func NewServerWithConfig(config Config) *Server {
	s := NewServer()
	if config.Port > 0 {
		s.port = config.Port
	}
	s.addStubs(config.Stubs...)
	s.loadStubs(config.StubsDir)
	return s
}

// Start starts the server.
func (s *Server) Start() {
	if s.dir == defaultStubsDir {
		s.loadStubs(defaultStubsDir)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/httpmock/add", s.addStubHandler)
	mux.HandleFunc("/httpmock/list", s.listStubsHandler)
	mux.HandleFunc("/", s.genericStubHandler)
	s.srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Info().Msgf("starting HTTP mock server on port %d", s.port)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msgf("failed to start HTTP mock server: %v", err)
		}
	}()
	time.Sleep(time.Second)
}

// Stop stops the server.
func (s *Server) Stop() error {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		log.Error().Msgf("failed to stop HTTP mock server: %v", err)
		return err
	}
	log.Info().Msgf("HTTP mock server stopped")
	return nil
}

// AddStub adds a stub to the server.
func (s *Server) AddStub(stub *Stub) {
	s.addStub(stub)
}

// AddStubs adds multiple stubs to the server.
func (s *Server) AddStubs(stubs ...*Stub) {
	s.addStubs(stubs...)
}

// WithStubs adds multiple stubs to the server.
func (s *Server) WithStubs(stubs ...*Stub) *Server {
	s.addStubs(stubs...)
	return s
}

// WithStubsFrom loads stubs from the given location.
func (s *Server) WithStubsFrom(stubsLocation string) *Server {
	s.loadStubs(stubsLocation)
	return s
}

// ClearStubs clears all stubs from the server.
func (s *Server) ClearStubs() {
	s.stubs = make(stubHashMap)
}

// WithPort sets the port for the server.
func (s *Server) WithPort(port int) *Server {
	s.port = port
	return s
}

// loadStubs loads stubs from the given location.
func (s *Server) loadStubs(location string) {
	s.dir = location
	files, err := os.ReadDir(location)
	if err != nil {
		log.Warn().Msgf("failed to read stubs directory: %s %v", location, err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			s.loadStubs(filepath.Join(location, file.Name()))
			continue
		}
		// only load .json and .yaml/.yml files
		if !strings.HasSuffix(file.Name(), ".json") && !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}
		fileBytes, err := os.ReadFile(fmt.Sprintf("%s/%s", location, file.Name()))
		if err != nil {
			log.Error().Msgf("failed to read stub file: %v", err)
			return
		}

		stub, err := getStubFromBytes(fileBytes)
		if err != nil {
			log.Error().Msgf("failed to parse stub file %s: %v", file.Name(), err)
			return
		}

		s.addStub(stub)
	}
}

// addStub adds a stub to the server.
func (s *Server) addStub(stub *Stub) []error {
	stub.Request.Sanitize()
	if errs := stub.validationErrors(); len(errs) > 0 {
		log.Error().Msgf("invalid stub: %v", errs)
		return errs
	}

	// compact JSON body before hashing
	if stub.Request.Body != nil {
		bodyBytes, err := json.Marshal(stub.Request.Body)
		if err != nil {
			log.Error().Msgf("failed to marshal stub body: %v", err)
			return []error{err}
		}
		buff := new(bytes.Buffer)
		if err = json.Compact(buff, bodyBytes); err != nil {
			log.Error().Msgf("failed to compact stub body: %v", err)
			return []error{err}
		}
		stub.Request.Body = buff.String()
	}

	h := getHash(stub.Request)
	if existing, ok := s.stubs[h]; ok {
		log.Warn().Msgf("overriding existing stub: %s ", existing.Request.String())
	}
	s.stubs[h] = stub
	log.Info().Msgf("added stub: %s", stub.Request.String())
	return []error{}
}

// addStubs adds multiple stubs to the server.
func (s *Server) addStubs(stubs ...*Stub) {
	for _, v := range stubs {
		s.addStub(v)
	}
}
