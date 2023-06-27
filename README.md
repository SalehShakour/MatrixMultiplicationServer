# Matrix Multiplication Server

Matrix Multiplication Server is a Go project that implements an HTTP server for performing matrix multiplication. The server can receive two matrices as input and return the result of their multiplication.

This repository contains two branches: one that implements the workers using RPC, and another that implements the workers using threads and semaphores.

## Branches

### Thread-Semaphore

The `Thread-Semaphore` branch contains an updated version of the project, which uses threads and semaphores instead of RPC. In this version, the server launches a specified number of worker threads and uses a semaphore to limit the number of active threads at any given time.

### RPC

The `RPC` branch contains the original implementation of the project, which uses RPC to communicate with the worker processes. In this version, the server launches a specified number of worker processes and uses RPC to distribute the multiplication tasks among them.


## Usage

To use the Matrix Multiplication Server, you need to send an HTTP POST request to the `/mulmatrix` endpoint with a JSON payload containing two matrices. The server will then perform the multiplication and return the result as a JSON array.

For example, here is a sample request that multiplies two matrices:

```sh
curl -X POST -H "Content-Type: application/json" -d '{"matrixA": [[1, 2], [3, 4]], "matrixB": [[5, 6], [7, 8]]}' http://localhost:9090/mulmatrix
```
or
```sh
Invoke-WebRequest -Uri http://localhost:9090/mulmatrix -Method POST -ContentType "application/json" -Body '{"matrixA": [[1, 2], [3, 4]], "matrixB": [[5, 6], [7, 8]]}'
```
And here is the response from the server:
```
[
    [19, 22],
    [43, 50]
]
```
You can also use the `/setnumworkers` endpoint to dynamically change the number of worker threads or processes. To do this, send an HTTP POST request to the `/setnumworkers` endpoint with the new number of workers as the request body.

For example, here is a sample request that sets the number of workers to 4:
```
curl -X POST -d '4' http://localhost:9090/setnumworkers
```
And here is the response from the server:
```
Number of workers set to 4
```


I hope this helps! Let me know if you have any further questions.
