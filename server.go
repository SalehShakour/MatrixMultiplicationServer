package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"sync"

	"golang.org/x/sync/semaphore"
)

type MatrixData struct {
	MatrixA [][]float64 `json:"matrixA"`
	MatrixB [][]float64 `json:"matrixB"`
}

// MyHandler A struct that implements the http.Handler interface
type MyHandler struct{}

// Matrix represents a 2D matrix of float64 values
type Matrix struct {
	Rows int
	Cols int
	Data [][]float64
}

// NewMatrix creates a new matrix with the given dimensions and fills it with random values
func NewMatrix(rows, cols int) *Matrix {
	m := &Matrix{
		Rows: rows,
		Cols: cols,
		Data: make([][]float64, rows, rows),
	}
	return m
}

// Print prints the matrix in a readable format
func (m *Matrix) Print() {
	for i := range m.Data {
		fmt.Println(m.Data[i])
	}
}

// MulMatrix multiplies two matrices using rpc calls to workers
func MulMatrix(a, b *Matrix) (*Matrix, error) {
	if a.Cols != b.Rows {
		return nil, fmt.Errorf("incompatible matrix dimensions: %dx%d and %dx%d", a.Rows, a.Cols, b.Rows, b.Cols)
	}
	c := NewMatrix(a.Rows, b.Cols)
	var wg sync.WaitGroup           // to wait for all workers to finish
	var mu sync.Mutex               // to protect access to the result matrix
	sem := semaphore.NewWeighted(4) // to limit the number of active workers

	// create a worker function that computes one row of the result matrix
	worker := func(i int) {
		defer wg.Done()      // decrement the wait group counter
		defer sem.Release(1) // release the semaphore

		client, err := rpc.DialHTTP("tcp", fmt.Sprintf("localhost:%d", 9000+i)) // connect to the worker rpc server
		if err != nil {
			log.Fatal(err)
		}

		args := &MulRowArgs{
			Row:  a.Data[i],
			Rows: b.Rows,
			Cols: b.Cols,
			Data: b.Data,
		} // create the arguments for the MulRow rpc method
		var reply []float64 // create a slice to store the reply from the MulRow rpc method
		fmt.Println("args ", *args)
		err = client.Call("Worker.MulRow", args, &reply) // call the MulRow rpc method and store the result in reply
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(reply)

		mu.Lock()
		if reply != nil {
			c.Data[i] = reply // copy the reply to the result matrix row
		}
		mu.Unlock()

	}

	for i := 0; i < a.Rows; i++ {
		wg.Add(1)           // increment the wait group counter
		sem.Acquire(nil, 1) // acquire the semaphore
		go worker(i)        // start the worker goroutine with the row index as argument

	}

	wg.Wait() // wait for all workers to finish

	return c, nil
}

// Worker is a struct that implements the rpc methods for matrix multiplication
type Worker struct{}

// MulRowArgs is a struct that holds the arguments for the MulRow rpc method
type MulRowArgs struct {
	Row  []float64 // the row vector to multiply by a matrix
	Rows int
	Cols int
	Data [][]float64
}

// MulRow is an rpc method that multiplies a row vector by a matrix and returns the result as a slice of float64
func (w *Worker) MulRow(args *MulRowArgs, reply *[]float64) error {
	res := make([]float64, args.Cols)
	for j := 0; j < args.Cols; j++ {
		sum := 0.0
		for k := 0; k < args.Rows; k++ {
			sum += args.Row[k] * args.Data[k][j]
		}
		res[j] = sum

	}
	*reply = res // set the reply pointer to point to the result slice

	return nil

}

// matrixHandler handles the requests for getting the matrices
func matrixHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method == "POST" {
		// Create an instance of MatrixData
		var data MatrixData

		// Read the request body and store it in a byte slice
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Parse the JSON data and store it in data
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Do something with data.MatrixA and data.MatrixB
		fmt.Fprintf(w, "Received matrix A: %v\n", data.MatrixA)
		fmt.Fprintf(w, "Received matrix B: %v\n", data.MatrixB)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// mulMatrixHandler handles the requests for multiplying the matrices
func mulMatrixHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method == "POST" {
		// Create an instance of MatrixData
		var data MatrixData

		// Read the request body and store it in a byte slice
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Parse the JSON data and store it in data
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Convert the slices to Matrix structs
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
		// Multiply the matrices using MulMatrix function
		c, err := MulMatrix(a, b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(c)

		// Convert the result matrix to a slice
		result := c.Data

		// Encode the result slice to JSON and send it to the response writer
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

// The ServeHTTP method that handles the requests
func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from MyHandler!")
}

func main() {
	// Create an instance of MatrixData with some sample data
	logger := log.New(os.Stdout, "", log.Ltime)

	myHandler := &MyHandler{}

	// Use a server mux to handle multiple routes
	mux := http.NewServeMux()

	// Register the handler functions with the mux
	mux.HandleFunc("/matrix", matrixHandler)       // Handle requests for getting the matrices
	mux.HandleFunc("/mulmatrix", mulMatrixHandler) // Handle requests for multiplying the matrices
	mux.Handle("/", myHandler)                     // Use the struct as a handler

	// Start 8 worker processes using exec.Command and exec.Start
	for i := 0; i < 8; i++ {
		go func(port int) {
			fmt.Println("port is ", 9000+port)
			cmd := exec.Command("F:\\Programmig\\Go\\HW2\\worker.exe", fmt.Sprintf("%d", 9000+port)) // create a command with the port number as argument
			stdout, _ := cmd.StdoutPipe()
			go handleSTDOUT(stdout)
			err := cmd.Start() // start the command as a new process
			if err != nil {
				logger.Fatal(err)
			}

			err = cmd.Wait()
			if err != nil {
				logger.Fatal(err)
			}
		}(i)
	}

	// Start the server and listen on port 9090
	logger.Println("Starting server on port 9090")
	err := http.ListenAndServe(":9090", mux)
	logger.Fatal(err)
}

func handleSTDOUT(closer io.ReadCloser) {
	b := make([]byte, 1000)
	for {
		nr, e := closer.Read(b)
		fmt.Println("from worker: ", string(b[:nr]))
		if e != nil {
			fmt.Println(e)
			return
		}
	}
}
