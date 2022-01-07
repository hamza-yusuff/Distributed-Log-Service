package server

// File responsiblef for implementing a JSON/HTTP web server
// Has two endpoints ->
// Produce for writing to the log
// Consume for reading from the log

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Handler Functions ->
// unmarshalls the request json body into the produce and consume structs
// runs the logic through the struct methods
// marshalls ans writes resutls to the response

// addr is the address on which the server would run
// returns the net/http server pointer
func NewHTTPServer(addr string) *http.Server {

	// returns an httpServer struct instace
	https := newHTTPServer()
	r := mux.NewRouter()

	// macthes the route to their handlers
	r.HandleFunc("/", https.handleProduce).Methods("POST")
	r.HandleFunc("/", https.handleConsume).Methods("GET")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

type httpServer struct {
	Log *Log
}

// similar to a constructor function, returns a pointer to the httpServer struct above
func newHTTPServer() *httpServer {
	return &httpServer{
		Log: NewLog(),
	}
}

// Struct where record is unmarshalled and write to log using the Handler
type ProduceRequest struct {
	Record Record `json:record`
}

// Struct where response record read's index from log is unmarshalled and sent
type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

// Struct where consumed request is unmarshalled for reading the record at Offset from log
type ConsumeRequest struct {
	Offset uint64
}

// Struct where consumed response is unmarshalled for sending the read record from the log
type ConsumeResponse struct {
	Record Record
}

// Main Handeler Functions below

// Method to the struct httpServer
// Implements the handleProduce functionality by -
// - unmarshalls request into struct
// - uses the struct to append record into the log
// - marshalls the results ( ProduceResponse struct) into the response
func (server *httpServer) handleProduce(write http.ResponseWriter, r *http.Request) {
	var req ProduceRequest
	// unmarshals the request body into the Produce Request struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(write, err.Error(), http.StatusBadRequest)
		return
	}
	off, err := server.Log.Append(req.Record)

	if err != nil {
		http.Error(write, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ProduceResponse{Offset: off}
	err = json.NewEncoder(write).Encode(res)
	if err != nil {
		http.Error(write, err.Error(), http.StatusInternalServerError)
		return
	}

}

// Does the same thing as handleProduce but uses Read to read from the log

func (server *httpServer) handleConsume(write http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(write, err.Error(), http.StatusBadRequest)
	}

	// reads fromt the log
	record, err := server.Log.Read(req.Offset)

	if err == ErrOffsetNotFound {
		http.Error(write, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(write, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConsumeResponse{Record: record}
	err = json.NewEncoder(write).Encode(res)

	if err != nil {
		http.Error(write, err.Error(), http.StatusInternalServerError)
		return
	}

}
