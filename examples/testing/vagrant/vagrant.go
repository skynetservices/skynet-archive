package main

import (
	"expvar"
	"flag"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var requests = flag.Int("requests", 10, "number of concurrent requests")
var doozer = flag.String("doozer", "127.0.0.1:8046", "doozer instance to connect to")

var totalRequests = expvar.NewInt("total-requests")
var successfulRequests = expvar.NewInt("successful-requests")

var testserviceClient *client.ServiceClient
var fibserviceClient *client.ServiceClient

func main() {
	flag.Parse()

	doozerConfig := &skynet.DoozerConfig{
		Uri:          *doozer,
		AutoDiscover: true,
	}

	clientConfig := &skynet.ClientConfig{
		DoozerConfig:       doozerConfig,
		ConnectionPoolSize: *requests,
		IdleTimeout:        (2 * time.Minute),
	}

	c := make(chan os.Signal, 1)
	quitChan := make(chan bool, 1)
	requestChan := make(chan string, *requests*3)
	workerQuitChan := make(chan bool, 1)
	workerWaitGroup := new(sync.WaitGroup)

	go watchSignals(c, quitChan)

	skynetClient := client.NewClient(clientConfig)
	testserviceClient = skynetClient.GetService("TestService", "", "", "")
	fibserviceClient = skynetClient.GetService("Fibonacci", "", "", "")

	fmt.Printf("Starting %d Workers\n", *requests)
	for i := 0; i < *requests; i++ {
		go worker(requestChan, workerWaitGroup, workerQuitChan)
	}

	requestNum := 0

	for {
		select {
		case <-quitChan:
			for i := 0; i < *requests; i++ {
				workerQuitChan <- true
			}

			workerWaitGroup.Wait()

			successful, _ := strconv.Atoi(successfulRequests.String())
			total, _ := strconv.Atoi(totalRequests.String())

			failed := total - successful

			percentSuccess := int(float64(successful) / float64(total) * 100)
			percentFailed := int(float64(failed) / float64(total) * 100)

			fmt.Printf("Total Requests: %d, Successful: %d (%d%%), Failed: %d (%d%%)\n", total, successful, percentSuccess, failed, percentFailed)
			return
		default:
			requestChan <- "testservice"

			requestNum++
		}
	}
}

func worker(requestChan chan string, waitGroup *sync.WaitGroup, quitChan chan bool) {
	waitGroup.Add(1)

	for {
		select {
		case <-quitChan:
			waitGroup.Done()
			return

		case service := <-requestChan:
			totalRequests.Add(1)

			switch service {
			case "testservice":
				fmt.Println("Sending TestService request")

				in := map[string]interface{}{
					"data": "Upcase me!!",
				}

				out := map[string]interface{}{}
				err := testserviceClient.Send(nil, "Upcase", in, &out)

				if err == nil && out["data"].(string) == "UPCASE ME!!" {
					successfulRequests.Add(1)
				}

			case "fibservice":
			}

		}
	}
}

func watchSignals(c chan os.Signal, quitChan chan bool) {
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM)

	for {
		select {
		case sig := <-c:
			switch sig.(syscall.Signal) {
			// Trap signals for clean shutdown
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM:
				quitChan <- true
				return
			}
		}
	}
}
