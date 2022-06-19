package main

import (
	"context"
	"flag"
	"fmt"

	"proj_2/token"
	"proj_2/utils"

	"google.golang.org/grpc"
)

func main() {

	// Accept command line args
	portPtr := flag.String("port", "50051",
		"port number where server is running")
	hostPtr := flag.String("host", "localhost",
		"host where server is running")

	createPtr := flag.Bool("create", false, "set to create token")
	dropPtr := flag.Bool("drop", false, "set to drop token")
	writePtr := flag.Bool("write", false, "set to write token")
	readPtr := flag.Bool("read", false, "set to read ptr")

	idPtr := flag.String("id", "undefined", "id of the token")
	namePtr := flag.String("name", "undefined", "name of the token")

	lowPtr := flag.Uint64("low", 1, "low value of the domain of token")
	midPtr := flag.Uint64("mid", 1, "mid value of the domain of token")
	highPtr := flag.Uint64("high", 1, "high value of the domain of token")

	readersPtr := flag.String("readers", "undefined",
		"space separated list of access points in the form <host>:<port>")
	writerPtr := flag.String("writer", "undefined",
		"single access point in the form <host>:<port>")

	flag.Parse()
	tail := flag.Args()

	// Create client connection
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", *hostPtr, *portPtr),
		grpc.WithInsecure())
	utils.IsSuccess(err)
	defer conn.Close()
	c := token.NewTokenServiceClient(conn)

	// Initialize blank request
	req := token.Request{}
	req.Domain = &token.Domain{}
	req.TokenState = &token.State{}

	// Initialize blank response
	resp := &token.Response{}

	// Fill in request, fire and get response
	if *createPtr {
		req.Id = *idPtr
		req.Writer = *writerPtr
		req.Readers = append([]string{*readersPtr}, tail...)
		resp, err = c.Create(context.Background(), &req)
	} else if *dropPtr {
		req.Id = *idPtr
		resp, err = c.Drop(context.Background(), &req)
	} else if *writePtr {
		req.Id = *idPtr
		req.Name = *namePtr
		req.Domain.Low = *lowPtr
		req.Domain.Mid = *midPtr
		req.Domain.High = *highPtr
		req.Requestip = *hostPtr
		req.Requestport = *portPtr
		resp, err = c.Write(context.Background(), &req)
	} else if *readPtr {
		req.Id = *idPtr
		req.Requestip = *hostPtr
		req.Requestport = *portPtr
		resp, err = c.Read(context.Background(), &req)
	}

	// Print server response
	fmt.Println("Server Response:", (resp.GetBody()))
	utils.IsSuccess(err)
}
