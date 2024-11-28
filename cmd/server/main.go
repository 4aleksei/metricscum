package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gaugeMetric float64
type counterMetric int64

type MetricsStorage interface {
	Add(string, gaugeMetric)
	Update(string, counterMetric)
}

type MemStorage struct {
	gauge   map[string]gaugeMetric
	counter map[string]counterMetric
}

func addGauge(storage MetricsStorage, name string, val gaugeMetric) {
	storage.Add(name, val)
}
func addCounter(storage MetricsStorage, name string, val counterMetric) {
	storage.Update(name, val)
}

func (storage MemStorage) Add(name string, val gaugeMetric) {
	storage.gauge[name] += val
}

func (storage MemStorage) Update(name string, val counterMetric) {
	storage.counter[name] = val
}

var stor MemStorage

func mainPageGauge(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	urlPart := strings.Split(req.URL.Path, "/")

	if len(urlPart) == 4 && urlPart[3] == "" {
		http.Error(res, "Bad data!", http.StatusNotFound)
		return
	}

	if len(urlPart) != 5 {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	/*	body := fmt.Sprintf("Method: %s\r\n", req.Method)
		body += "Header ===============\r\n"
		for k, v := range req.Header {
			body += fmt.Sprintf("%s: %v\r\n", k, v)
		}
	*/
	/*body += "Query parameters ===============\r\n"
	if err := req.ParseForm(); err != nil {
		res.Write([]byte(err.Error()))
		return
	}
	for k, v := range req.Form {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	body += fmt.Sprintf("%s\r\n", req.URL)
	*/
	value, err := strconv.ParseFloat(urlPart[4], 64)
	if err != nil {
		http.Error(res, "Bad gauge value!", http.StatusBadRequest)
		return
	}
	addGauge(stor, urlPart[3], gaugeMetric(value))

}

func mainPageCounter(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	urlPart := strings.Split(req.URL.Path, "/")

	if len(urlPart) == 4 && urlPart[3] == "" {
		http.Error(res, "Bad data!", http.StatusNotFound)
		return
	}

	if len(urlPart) != 5 {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	/*	body := fmt.Sprintf("Method: %s\r\n", req.Method)
		body += "Header ===============\r\n"
		for k, v := range req.Header {
			body += fmt.Sprintf("%s: %v\r\n", k, v)
		}
		body += "Query parameters ===============\r\n"
		if err := req.ParseForm(); err != nil {
			res.Write([]byte(err.Error()))
			return
		}
		for k, v := range req.Form {
			body += fmt.Sprintf("%s: %v\r\n", k, v)
		}
		body += fmt.Sprintf("%s\r\n", req.URL)

		body += fmt.Sprintf("%s\r\n", urlPart[4])
	*/
	value, err := strconv.ParseInt(urlPart[4], 10, 64)
	if err != nil {
		http.Error(res, "Bad counter value!", http.StatusBadRequest)
		return
	}
	addCounter(stor, urlPart[3], counterMetric(value))
	//	res.Write([]byte(body))

}

func mainPageError(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	http.Error(res, "Bad request", http.StatusBadRequest)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc(`/update/gauge/`, mainPageGauge)
	mux.HandleFunc(`/update/counter/`, mainPageCounter)
	mux.HandleFunc(`/update/`, mainPageError)
	//mux.HandleFunc(`/`, mainPageError)

	if stor.counter == nil {
		stor.counter = make(map[string]counterMetric)
	}
	if stor.gauge == nil {
		stor.gauge = make(map[string]gaugeMetric)
	}

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
