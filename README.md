# goRPC

Minimalist RPC framework for Go. Write your service definitions in Go and generate anything from it.

Currently, goRPC provides client templates for:
* Go
* PHP
* TypeScript
* Swift
* JavaScript
* Python

## Quick start
This guide will get you started with goRPC by providing a simple working example.
You will learn how to make server and client using goRPC.

### Prerequisites
* Go v1.13 or newer. For installation instructions, see [Go’s Getting Started](https://golang.org/doc/install) guide.

### Install tool
```shell
go install github.com/damejeras/gorpc@latest
```

### Initiate project
Create a project and get `transport` library:
```shell
mkdir gorpc-example
cd gorpc-example
go mod init gorpc-example
go get github.com/damejeras/gorpc/transport
```

### Create service definition
Create `definition/greeter.go` with service definition:
```go
package definition

// GreeterService is service definition.
type GreeterService interface {
  // SayHello ends a greeting
  SayHello(HelloRequest) HelloResponse
}

// HelloRequest message containing the user's name.
type HelloRequest struct {
  Name string
}

// HelloResponse message containing the greetings
type HelloResponse struct {
  Greeting string
}
```

### Generate server interface
Fetch server template:
```shell
wget https://raw.githubusercontent.com/damejeras/gorpc/main/templates/server.go.tmpl
```
Generate server interface and format it with `gofmt`:
```shell
gorpc --template=server.go.tmpl --package main definition/greeter.go --output server.go
gofmt -w server.go server.go
```
By now you should have `server.go` containing `GreeterService` interface, `HelloRequest` and `HelloResponse` structs.

### Write server implementation
1. Create `main.go`:
```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/damejeras/gorpc/transport"
)

type greeterService struct{}

func (g greeterService) SayHello(ctx context.Context, request HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{
		Greeting: "Hello " + request.Name,
	}, nil
}

func main() {
	server := transport.NewServer()
	RegisterGreeterService(server, greeterService{})

	if err := http.ListenAndServe(":8000", server); err != nil {
		log.Fatal(err)
	}
}
```
2. Run server with `go run .`.

### Generate client code
1. Make client package directory:
```shell
mkdir client
```
2. Fetch client's template:
```shell
wget https://raw.githubusercontent.com/damejeras/gorpc/main/templates/client.go.tmpl
```
3. Generate client code:
```shell
gorpc --template=client.go.tmpl --package main definition/greeter.go --output client/client.go
gofmt -w client/client.go client/client.go
```

### Test your client
To test generated client create `client/main.go`:
```go
package main

import (
	"context"
	"fmt"
)

func main() {
	client := New("http://localhost:8000/")
	service := NewGreeterService(client)
	resp, err := service.SayHello(context.Background(), HelloRequest{Name: "Joe"})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Greeting)
}
```

Run client with:
```shell
go run ./client
Hello Joe
```


## Contributions

`goRPC` is a fork of https://github.com/pacedotdev/oto. Thank you to all developers that brought this fantastic project to the world.
