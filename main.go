package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yfuruyama/crzerolog"
	"go.opencensus.io/plugin/ochttp"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/serial", serialHandler).Methods("GET")
	r.HandleFunc("/go", goHandler).Methods("GET")
	r.HandleFunc("/go-array", goArrayHandler).Methods("POST")

	rootLogger := zerolog.New(os.Stdout)
	middleware := crzerolog.InjectLogger(&rootLogger)

	handler := cors.Default().Handler(r)
	handler = middleware(handler)

	httpHandler := &ochttp.Handler{
		Propagation: &propagation.HTTPFormat{},
		Handler:     handler,
	}

	log.Info().Msg("Starting server...")

	if err := http.ListenAndServe(":8080", httpHandler); err != nil {
		log.Fatal().Err(err).Msg("Can't start server")
	}

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	logger.Info().Msg("Request on Health checker")

	time.Sleep(1 * time.Second)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func serialHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	logger.Info().Msg("Request on Serial")

	res1 := makeRequest("http://localhost:8080/")
	res2 := makeRequest("http://localhost:8080/")
	res3 := makeRequest("http://localhost:8080/")

	fmt.Println(res1)
	fmt.Println(res2)
	fmt.Println(res3)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func makeRequest(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return ":("
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ":("
	}

	return string(body)
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	logger.Info().Msg("Request on Go")

	c := make(chan string)

	go makeRequestWithChannel("http://localhost:8080/", c)
	go makeRequestWithChannel("http://localhost:8080/", c)
	go makeRequestWithChannel("http://localhost:8080/", c)

	fmt.Println(<-c)
	fmt.Println(<-c)
	fmt.Println(<-c)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func makeRequestWithChannel(url string, c chan string) {
	res, err := http.Get(url)
	if err != nil {
		c <- ":("
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c <- ":("
	}

	c <- string(body)
}

func goArrayHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	logger.Info().Msg("Request on Go-Array")

	// Getting JSON request
	gar := GoArrayRequest{}
	err := json.NewDecoder(r.Body).Decode(&gar)
	if err != nil {
		logger.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Goroutines
	c := make(chan string)
	for _, url := range gar.URLs {
		go makeRequestWithChannel(url, c)
	}
	for _, _ = range gar.URLs {
		fmt.Println(<-c)
	}

	// Response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// GoArrayRequest body JSON
type GoArrayRequest struct {
	URLs []string `json:"urls"`
}
