package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var (
	numWorkers = 8
	sem        = semaphore.NewWeighted(int64(numWorkers))
)

type MatrixData struct {
	MatrixA [][]float64 `json:"matrixA"`
	MatrixB [][]float64 `json:"matrixB"`
}

type MyHandler struct{}

type Matrix struct {
	Rows int
	Cols int
	Data [][]float64
}

func NewMatrix(rows, cols int) *Matrix {
	m := &Matrix{
		Rows: rows,
		Cols: cols,
		Data: make([][]float64, rows, rows),
	}
	return m
}

func (m *Matrix) Print() {
	for i := range m.Data {
		fmt.Println(m.Data[i])
	}
}

func setNumWorkersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		newNumWorkers, err := strconv.Atoi(string(body))
		if err != nil || newNumWorkers < 1 || newNumWorkers > 64 {
			http.Error(w, "Invalid value for numWorkers", http.StatusBadRequest)
			return
		}

		numWorkers = newNumWorkers
		sem = semaphore.NewWeighted(int64(numWorkers))

		fmt.Fprintf(w, "Number of workers set to %d\n", newNumWorkers)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func MulMatrix(a, b *Matrix, numWorkers int) (*Matrix, error) {
	if a.Cols != b.Rows {
		return nil, fmt.Errorf("incompatible matrix dimensions: %dx%d and %dx%d", a.Rows, a.Cols, b.Rows, b.Cols)
	}
	c := NewMatrix(a.Rows, b.Cols)
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := semaphore.NewWeighted(int64(numWorkers))

	worker := func(i int) {
		defer wg.Done()
		defer sem.Release(1)

		res := make([]float64, b.Cols)
		for j := 0; j < b.Cols; j++ {
			sum := 0.0
			for k := 0; k < b.Rows; k++ {
				sum += a.Data[i][k] * b.Data[k][j]
			}
			res[j] = sum
		}

		mu.Lock()
		c.Data[i] = res
		mu.Unlock()
	}

	for i := 0; i < a.Rows; i++ {
		wg.Add(1)
		sem.Acquire(context.Background(), 1)
		go worker(i)
	}

	wg.Wait()

	return c, nil
}

func matrixHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var data MatrixData

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Received matrix A: %v\n", data.MatrixA)
		fmt.Fprintf(w, "Received matrix B: %v\n", data.MatrixB)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func mulMatrixHandler(numWorkers int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var data MatrixData

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			err = json.Unmarshal(body, &data)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			a := &Matrix{
				Rows: len(data.MatrixA),
				Cols: len(data.MatrixA[0]),
				Data: data.MatrixA,
			}

			b := &Matrix{
				Rows: len(data.MatrixB),
				Cols: len(data.MatrixB[0]),
				Data: data.MatrixB,
			}

			fmt.Println(a)
			fmt.Println(b)
			c, err := MulMatrix(a, b, numWorkers)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println(c)

			result := c.Data

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(result)
			if err != nil {
				http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func main() {
	myHandler := &MyHandler{}

	mux := http.NewServeMux()

	numWorkers := 8

	mux.HandleFunc("/matrix", matrixHandler)
	mux.HandleFunc("/mulmatrix", mulMatrixHandler(numWorkers))
	mux.Handle("/", myHandler)
	mux.HandleFunc("/setnumworkers", setNumWorkersHandler)

	log.Println("Starting server on port 9090")
	err := http.ListenAndServe(":9090", mux)
	log.Fatal(err)
}
