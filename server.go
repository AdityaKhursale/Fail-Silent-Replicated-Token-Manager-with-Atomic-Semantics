package main

import (
	"flag"
	"fmt"
	"net"

	"proj_2/token"
	"proj_2/utils"

	"google.golang.org/grpc"
)

func main() {

	// Accept command line args
	hostPtr := flag.String("host", "localhost", "host where server is running")
	portPtr := flag.String("port", "50051", "port number to use")
	flag.Parse()

	// Start Server
	fmt.Println("\nServer started on host:port ", (*hostPtr), (*portPtr), "\n")
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", (*hostPtr), (*portPtr)))
	utils.IsSuccess(err)
	s := token.Server{}
	server := grpc.NewServer()
	token.RegisterTokenServiceServer(server, &s)
	err = server.Serve(ln)
	utils.IsSuccess(err)
}
