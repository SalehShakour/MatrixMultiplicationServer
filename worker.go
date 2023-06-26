package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

// Worker is a struct that implements the rpc methods for matrix multiplication
type Worker struct{}

// MulRowArgs is a struct that holds the arguments for the MulRow rpc method
type MulRowArgs struct {
	Row    []float64 // the row vector to multiply by a matrix
	Matrix *Matrix   // the matrix to multiply by the row vector
}

// Matrix represents a 2D matrix of float64 values
type Matrix struct {
	rows int
	cols int
	data [][]float64
}

// MulRow is an rpc method that multiplies a row vector by a matrix and returns the result as a slice of float64
func (w *Worker) MulRow(args *MulRowArgs, reply *[]float64) error {
	res := make([]float64, args.Matrix.cols)
	for j := 0; j < args.Matrix.cols; j++ {
		sum := 0.0
		for k := 0; k < args.Matrix.rows; k++ {
			sum += args.Row[k] * args.Matrix.data[k][j]
		}
		res[j] = sum

	}
	*reply = res // set the reply pointer to point to the result slice

	return nil

}

func main() {
	// Create an instance of Worker
	worker := new(Worker)

	// Register the Worker with rpc
	err := rpc.Register(worker)
	if err != nil {
		log.Fatal(err)
	}

	// Register a HTTP handler for rpc
	rpc.HandleHTTP()

	// Get the port number from the command line argument
	port := os.Args[1]

	// Listen on the given port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	// Start serving rpc requests over HTTP
	fmt.Println("Worker listening on port", port)
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal(err)
	}
}
