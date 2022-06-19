package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"proj_2/utils"

	"gopkg.in/yaml.v3"
)

type TokenConfig struct {
	Token   int
	Writer  string
	Readers []string
}

type TokenAccessPoints struct {
	Readers []string
	Writer  string
}

type serverNodes map[string]struct{}

func (s serverNodes) add(node string) {
	s[node] = struct{}{}
}

func (s serverNodes) delete(node string) {
	delete(s, node)
}

func (s serverNodes) has(node string) bool {
	_, exists := s[node]
	return exists
}

// Function to stop all launched servers
func cleanup(cmdHndls []exec.Cmd) {
	fmt.Println()
	fmt.Println("*** cleaning up...")
	for _, cmdHndl := range cmdHndls {
		cmdHndl.Process.Kill()
	}
	fmt.Println("stopped all servers")
	fmt.Println()
}

func main() {

	fPtr := flag.String("yaml", "configuration.yaml",
		"yaml file name containing replication configuratation")
	flag.Parse()

	var cmdHndls []exec.Cmd

	// Ctrl + C handler - lauch cleanup
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(cmdHndls)
		os.Exit(0)
	}()

	// Read initial configuration from YAML
	// Create tokenmap and nodeset s.t.
	// tokenmap is of form {<token_id> : TokenAccessPoints(Readers, Writers)}
	// nodeset is list of all unique server access points
	fmt.Println("***Reading YAML")
	configFile, err := ioutil.ReadFile(*fPtr)
	utils.IsSuccess(err)

	var config []TokenConfig

	err = yaml.Unmarshal(configFile, &config)
	utils.IsSuccess(err)

	tokenMap := make(map[int]TokenAccessPoints)
	nodeSet := serverNodes{}

	for _, tokenData := range config {
		var tokenMapVal TokenAccessPoints
		tokenMapVal.Writer = tokenData.Writer
		tokenMapVal.Readers = tokenData.Readers
		if !nodeSet.has(tokenData.Writer) {
			nodeSet.add(tokenData.Writer)
		}
		for _, reader := range tokenData.Readers {
			if !nodeSet.has(reader) {
				nodeSet.add(reader)
			}
		}
		tokenMap[tokenData.Token] = tokenMapVal
	}

	// spawn all servers
	fmt.Println()
	fmt.Println("***Launching all servers")
	for node, _ := range nodeSet {
		fmt.Println("\tLaunching server on", node)
		ap := strings.Split(node, ":")
		cmd := fmt.Sprintf(
			"go run server.go -host %s -port %s"+
				" > output/server_op_%s_%s.txt 2>&1 &", ap[0], ap[1], ap[0], ap[1])
		fmt.Println("\tCommand: ", cmd)
		cmdHndl := exec.Command("bash", "-c", cmd)
		cmdHndl.Start()
		time.Sleep(1 * time.Second)
		cmdHndls = append(cmdHndls, *cmdHndl)
	}

	/*
		Create all initial tokens and their replicas
		- As per project description this is static hence
		  creating this way, and it is assumed no new tokens are created/dropped
		  except this initial configuration
	*/

	fmt.Println()
	fmt.Println("***Setting up all tokens with replicas")
	for tokenId, aps := range tokenMap {
		ap := strings.Split(aps.Writer, ":")
		readers := strings.Join(aps.Readers, " ")

		cmd := fmt.Sprintf(
			"go run client.go -create -id %d -host %s -port %s"+
				" -writer %s -readers %s",
			tokenId, ap[0], ap[1], aps.Writer, readers)
		fmt.Println("\tCommand: ", cmd)

		op, _ := exec.Command("bash", "-c", cmd).Output()
		fmt.Print("\t" + string(op))

		for _, node := range aps.Readers {
			if node == aps.Writer {
				continue
			}
			ap := strings.Split(node, ":")
			readers := strings.Join(aps.Readers, " ")
			cmd := fmt.Sprintf(
				"go run client.go -create -id %d -host %s -port %s"+
					" -writer %s -readers %s",
				tokenId, ap[0], ap[1], aps.Writer, readers)
			fmt.Println("\tCommand: ", cmd)
			op, _ := exec.Command("bash", "-c", cmd).Output()
			fmt.Print("\t" + string(op))
		}
	}
	fmt.Println("Intial replication is complete!")

	fmt.Println()
	fmt.Println("***System is up and available to test read and write requests")
	fmt.Println("\tPress Ctrl + C to close all servers")
	for {
		time.Sleep(10 * time.Second)
	}
}
