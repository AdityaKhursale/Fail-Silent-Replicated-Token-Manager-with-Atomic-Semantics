package token

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"proj_2/utils"

	"google.golang.org/grpc"
)

type Server struct {
	UnimplementedTokenServiceServer
}

type TokenType struct {
	request   *Request
	mutex     *sync.Mutex
	timestamp string
}

var tokenStore = make(map[string](TokenType))
var queryNumber uint64 = 0

// Function to get the token contents
func GetTokenState(t *Request) string {
	tStr := fmt.Sprintf("{id: %s, name: %s, "+
		"domain: {low: %d, mid: %d, high: %d}, "+
		"state: {partial_val: %d, final_val: %d}, "+
		"writer: %s, "+
		"readers: %s}",
		t.GetId(), t.GetName(), t.GetDomain().GetLow(),
		t.GetDomain().GetMid(), t.GetDomain().GetHigh(),
		t.GetTokenState().GetPartialval(), t.GetTokenState().GetFinalval(),
		t.GetWriter(), t.GetReaders())
	return tStr
}

/*
	Since the project description says nothing about the specifics of
	Create request Furthermore, states that we have to assume all Tokens
	are reflected at all reader and writer nodes,
	I did not implemnt replication scheme at this Create API level.
	Hence, calling create directly on Server through client without
	modifying launcher will not work as expected and
	will just create token on the server which it was requested
*/

// Function to create token - Same as of Project 2
func (s *Server) Create(ctx context.Context, req *Request) (*Response, error) {
	queryNumber += 1
	currQueryNumber := queryNumber

	reqQuery := fmt.Sprintf("{Action: create, Id: %s}", (req.GetId()))
	fmt.Println("\n** Request Received --> Request Number: ",
		currQueryNumber, ", Request: ", reqQuery)

	_, keyExists := tokenStore[req.GetId()]
	if keyExists {
		fmt.Println("Processed request number ",
			currQueryNumber, ", error occured")
		return &Response{Body: "Token is already present"},
			errors.New("token creation failed")
	}

	tokenStore[req.GetId()] = TokenType{request: req, mutex: &sync.Mutex{}}

	fmt.Println("Processed request number ", currQueryNumber,
		", Token State Now: ", GetTokenState(tokenStore[req.GetId()].request))
	fmt.Println("Tokenstore contains: ", reflect.ValueOf(tokenStore).MapKeys())

	return &Response{Body: fmt.Sprintf("Token created with id: %s",
		(req.GetId()))}, nil
}

// Function to drop token - Same as of Project 2
// Assumptions for this API is similar to that of Create API
func (s *Server) Drop(ctx context.Context, req *Request) (*Response, error) {
	queryNumber += 1
	currQueryNumber := queryNumber

	reqQuery := fmt.Sprintf("{Action: drop, Id: %s}", (req.GetId()))
	fmt.Println("\n** Request Received --> Request Number: ",
		currQueryNumber, ", Request: ", reqQuery)

	_, keyExists := tokenStore[req.GetId()]
	if !keyExists {
		fmt.Println("Processed request number ",
			currQueryNumber, ", error occured")
		return &Response{Body: "Token is absent, nothing to delete"},
			errors.New("token drop failed")
	}

	tokenStore[req.GetId()].mutex.Lock()
	defer tokenStore[req.GetId()].mutex.Unlock()
	delete(tokenStore, req.GetId())

	fmt.Println("Processed request number ", currQueryNumber,
		", Token State Now: ", GetTokenState(tokenStore[req.GetId()].request))
	fmt.Println("Tokenstore contains: ", reflect.ValueOf(tokenStore).MapKeys())

	return &Response{Body: fmt.Sprintf("Token dropped with id: %s",
		(req.GetId()))}, nil
}

// Function to serve reader's broadcast request
func (s *Server) ReadBroadcast(ctx context.Context,
	breq *ReadBroadcastRequest) (*ReadBroadcastResponse, error) {

	replica, exists := tokenStore[breq.GetTokenid()]
	if exists {

		return &ReadBroadcastResponse{
			Tokenid:    replica.request.GetId(),
			Domain:     replica.request.GetDomain(),
			TokenState: replica.request.GetTokenState(),
			Timestamp:  replica.timestamp,
			Status:     true}, nil
	}
	return &ReadBroadcastResponse{
		Status: false}, nil
}

// Function to write serve writer's broadcast request
func (s *Server) WriteBroadcast(ctx context.Context,
	breq *WriteBroadcastRequest) (*WriteBrodcastResponse, error) {

	tokenStore[breq.GetTokenid()].mutex.Lock()
	defer tokenStore[breq.GetTokenid()].mutex.Unlock()

	replica, exists := tokenStore[breq.GetTokenid()]
	if exists {
		if breq.Timestamp >= replica.timestamp || breq.Isreading {
			replica.timestamp = breq.GetTimestamp()
			replica.request.Domain = breq.GetDomain()
			replica.request.TokenState = breq.GetTokenState()
			tokenStore[breq.GetTokenid()] = replica
			return &WriteBrodcastResponse{Status: true}, nil
		}
	}
	return &WriteBrodcastResponse{Status: false}, nil
}

// Function to send broadcast request of write
func paralleWriteBroadcast(reader string,
	breq *WriteBroadcastRequest, ch chan int) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("\t\t----: " + reader + " is not available :----")
		}
	}()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(reader, grpc.WithInsecure())
	if err == nil {

		c := NewTokenServiceClient(conn)
		defer conn.Close()

		resp, err := c.WriteBroadcast(context.Background(), breq)
		utils.IsSuccess(err)

		fmt.Println("\t\t"+reader+"'s response: ", resp.GetStatus())

		if resp.GetStatus() {
			ch <- 1
		}
	}
}

// Function to send broadcast request of read
func parallelReadBroadcast(reader string,
	breq *ReadBroadcastRequest, ch chan int,
	readerTimestamps map[string]ReadBroadcastResponse) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("\t\t----: " + reader + " is not available :----")
		}
	}()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(reader, grpc.WithInsecure())
	if err == nil {
		c := NewTokenServiceClient(conn)
		defer conn.Close()

		resp, err := c.ReadBroadcast(context.Background(), breq)
		utils.IsSuccess(err)

		fmt.Println("\t\t"+reader+"'s response timestamp: ", resp.GetTimestamp(),
			" and status: ", resp.GetStatus())

		readerTimestamps[reader] = *resp
		if resp.GetStatus() {
			ch <- 1
		}
	}
}

// Function to write token - Changed for Project 3
func (s *Server) Write(ctx context.Context, req *Request) (*Response, error) {
	queryNumber += 1
	currQueryNumber := queryNumber

	// Log query
	reqQuery := fmt.Sprintf(
		"{Action: write, Id: %s, Name: %s, Low: %d, Mid: %d, High: %d}",
		req.GetId(), req.GetName(), req.GetDomain().GetLow(),
		req.GetDomain().GetMid(), req.GetDomain().GetHigh())
	fmt.Println("\n** Request Received --> Request Number: ",
		currQueryNumber, ", Request: ", reqQuery)

	// Check if token is available
	val, keyExists := tokenStore[req.GetId()]
	if !keyExists {
		fmt.Println("Processed request number ", currQueryNumber,
			", error occured")
		return &Response{Body: "Token is not available"}, nil
	}

	// Check if write previliges is available for this token to this server
	if val.request.GetWriter() != req.GetRequestip()+":"+req.GetRequestport() {
		fmt.Println("Processed request number ", currQueryNumber,
			", error occured")
		return &Response{Body: "No write previliges for this token"}, nil
	}

	// Acquire lock
	tokenStore[req.GetId()].mutex.Lock()
	defer tokenStore[req.GetId()].mutex.Unlock()

	// Calculate state and current timestamp i.e. partial value and final value
	partialVal := utils.FindArgminxHash(
		req.GetName(), req.GetDomain().GetLow(),
		req.GetDomain().GetMid())

	/*
		Shifted final value calculation from Read API to here
		Without this readers' are also writer and then project description
		does not make sense to me, Discussed with TA
	*/
	var finalVal uint64
	minMidHigh := utils.FindArgminxHash(
		req.GetName(), req.GetDomain().GetMid(),
		req.GetDomain().GetHigh())
	if minMidHigh < partialVal {
		finalVal = minMidHigh
	} else {
		finalVal = partialVal
	}

	currTimeStamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	currState := &State{Partialval: partialVal, Finalval: finalVal}
	fmt.Println("\tCalculated state of the token"+
		", state: {partial_val:", partialVal, " final_val:", finalVal, "}")

	// Create write broadcast request
	breq := WriteBroadcastRequest{}
	breq.Tokenid = req.GetId()
	breq.Domain = req.GetDomain()
	breq.Timestamp = currTimeStamp
	breq.TokenState = currState
	breq.Isreading = false

	// Broadcast write request
	fmt.Println("\tBrodcasting write requests")
	ch := make(chan int, 1)
	readers := tokenStore[req.GetId()].request.GetReaders()
	for _, reader := range readers {
		if reader == req.GetRequestip()+":"+req.GetRequestport() {
			continue
		}
		fmt.Println("\t\tBroadcast request sent to", reader)
		go paralleWriteBroadcast(reader, &breq, ch)
	}

	// Check majority
	acks := 0
	for {
		<-ch
		acks += 1
		if (acks) >= int((len(readers)/2)+1) {
			fmt.Println("\tGot the majority!")
			fmt.Println("\tWrite success on ", acks, "servers")
			break
		}
	}

	val.request.Name = req.GetName()
	val.request.Domain = req.GetDomain()
	val.request.TokenState = currState
	val.timestamp = currTimeStamp
	tokenStore[req.GetId()] = val

	fmt.Println("Processed request number ", currQueryNumber,
		", Token State Now: ", GetTokenState(tokenStore[req.GetId()].request))
	fmt.Println("Tokenstore contains: ", reflect.ValueOf(tokenStore).MapKeys())

	return &Response{Body: fmt.Sprintf(
		"Token updated with state: {partial_val: %d, final_val: %d}",
		tokenStore[req.GetId()].request.GetTokenState().GetPartialval(),
		tokenStore[req.GetId()].request.GetTokenState().GetFinalval())}, nil
}

func (s *Server) Read(ctx context.Context, req *Request) (*Response, error) {
	queryNumber += 1
	currQueryNumber := queryNumber

	// Log query
	reqQuery := fmt.Sprintf("{Action: read, Id: %s}", (req.GetId()))
	fmt.Println("\n** Request Received --> Request Number: ",
		currQueryNumber, ", Request: ", reqQuery)

	// Check if token is available
	val, keyExists := tokenStore[req.GetId()]
	if !keyExists {
		fmt.Println("Processed request number ", currQueryNumber,
			", error occured")
		return &Response{Body: "Token is not available"},
			nil
	}

	// Create reader broadcast request
	breq := ReadBroadcastRequest{}
	breq.Tokenid = req.GetId()

	// Acquire lock
	tokenStore[req.GetId()].mutex.Lock()
	defer tokenStore[req.GetId()].mutex.Unlock()

	readers := tokenStore[req.GetId()].request.GetReaders()

	// Check if read previliges is available for this token to this server
	readerPreviliged := false
	for _, reader := range readers {
		if reader == req.GetRequestip()+":"+req.GetRequestport() {
			readerPreviliged = true
		}
	}
	if !readerPreviliged {
		fmt.Println("Processed request number ", currQueryNumber,
			", error occured")
		return &Response{Body: "No read previliges for this token"}, nil
	}

	// Broadcast read request
	fmt.Println("\tBrodcasting read requests")
	readerTimestamps := make(map[string]ReadBroadcastResponse)
	ch := make(chan int, 1)
	for _, reader := range readers {
		if reader == req.GetRequestip()+":"+req.GetRequestport() {
			continue
		}
		fmt.Println("\t\tBroadcast request sent to", reader)
		go parallelReadBroadcast(reader, &breq, ch, readerTimestamps)
	}

	acks := 0
	for {
		<-ch
		acks += 1
		if (acks) >= int((len(readers)/2)+1) {
			fmt.Println("\tGot the majority!")
			break
		}
	}

	// Finding reader with maximum reader timestamp
	var maxReaderResponse ReadBroadcastResponse
	var maxTimestamp string
	var maxReader string

	fmt.Println("\tFinding the reader with highest timestamp")
	for reader, v := range readerTimestamps {
		if v.GetTimestamp() >= maxTimestamp {
			maxTimestamp = v.GetTimestamp()
			maxReaderResponse = v
			maxReader = reader
		}
	}
	fmt.Println("\tReader with highest timestamp is: " + maxReader)

	// Write back requests with brodcast
	wbreq := WriteBroadcastRequest{}
	wbreq.Tokenid = req.GetId()
	wbreq.Domain = maxReaderResponse.GetDomain()
	wbreq.Timestamp = maxReaderResponse.GetTimestamp()
	wbreq.TokenState = maxReaderResponse.GetTokenState()
	wbreq.Isreading = true

	fmt.Println("\tBrodcasting write-back requests based on values of", maxReader)
	for _, reader := range readers {
		if reader == req.GetRequestip()+":"+req.GetRequestport() {
			continue
		}
		fmt.Println("\t\tBroadcast request sent to", reader)
		go paralleWriteBroadcast(reader, &wbreq, ch)
	}

	acks = 0
	for {
		<-ch
		acks += 1
		if (acks) >= int((len(readers)/2)+1) {
			fmt.Println("\tGot the majority!")
			fmt.Println("\tWrite success on ", acks, "servers")
			break
		}
	}

	val.request.Domain = maxReaderResponse.GetDomain()
	val.request.TokenState = maxReaderResponse.GetTokenState()
	val.timestamp = maxReaderResponse.GetTimestamp()
	tokenStore[req.GetId()] = val

	fmt.Println("Processed request number ", currQueryNumber,
		", Token State Now: ", GetTokenState(tokenStore[req.GetId()].request))
	fmt.Println("Tokenstore contains: ", reflect.ValueOf(tokenStore).MapKeys())

	return &Response{Body: fmt.Sprintf(
		"Token updated with state: {partial_val: %d, final_val: %d}",
		tokenStore[req.GetId()].request.GetTokenState().GetPartialval(),
		tokenStore[req.GetId()].request.GetTokenState().GetFinalval())}, nil
}
