package main

import (
	"expvar"
	"flag"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/examples/testing/fibonacci"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var requests = flag.Int("requests", 10, "number of concurrent requests")
var doozer = flag.String("doozer",
	skynet.GetDefaultEnvVar("SKYNET_DZHOST", "127.0.0.1:8046"),
	"doozer instance to connect to")

var totalRequests = expvar.NewInt("total-requests")
var successfulRequests = expvar.NewInt("successful-requests")

var testserviceClient client.ServiceClient
var fibserviceClient client.ServiceClient

func main() {
	flag.Parse()

	doozerConfig := &skynet.DoozerConfig{
		Uri:          *doozer,
		AutoDiscover: true,
	}

	clientConfig := &skynet.ClientConfig{
		DoozerConfig:              doozerConfig,
		IdleConnectionsToInstance: *requests,
		MaxConnectionsToInstance:  *requests,
		IdleTimeout:               (2 * time.Minute),
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

	startTime := time.Now().UnixNano()
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
			stopTime := time.Now().UnixNano()

			successful, _ := strconv.Atoi(successfulRequests.String())
			total, _ := strconv.Atoi(totalRequests.String())

			failed := total - successful

			percentSuccess := int(float64(successful) / float64(total) * 100)
			percentFailed := int(float64(failed) / float64(total) * 100)

			runtime := (stopTime - startTime) / 1000000
			rqps := float64(total) / (float64(runtime) / 1000)

			fmt.Printf("======================================")
			fmt.Printf("======================================")
			fmt.Printf("Completed in %d Milliseconds, %f Requests/s\n",
				runtime, rqps)
			fmt.Printf("\nTotal Requests: %d, Successful: %d (%d%%)",
				total, successful, percentSuccess)
			fmt.Printf(", Failed: %d (%d%%)\n\n", failed, percentFailed)
			return

		default:
			if requestNum%2 == 0 {
				requestChan <- "testservice"
			} else {
				requestChan <- "fibservice"
			}

			requestNum++
		}
	}
}

func worker(requestChan chan string, waitGroup *sync.WaitGroup,
	quitChan chan bool) {

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

				randString := strconv.FormatUint(uint64(rand.Uint32()), 35)
				randString = randString + randString + randString

				in := map[string]interface{}{
					"data": randString,
				}

				fmt.Println("Sending TestService request: " + in["data"].(string))

				out := map[string]interface{}{}
				err := testserviceClient.Send(nil, "Upcase", in, &out)

				upper := strings.ToUpper(randString)
				if err == nil && out["data"].(string) == upper {
					successfulRequests.Add(1)
					fmt.Println("TestService returned: " + out["data"].(string))
				}

			case "fibservice":
				req := fibonacci.Request{
					Index: rand.Intn(10),
				}

				// It's possible that rand could have returned 0, and
				// we are using that as our blank Value let's set it
				// to something else when 0 happens to get selected
				if req.Index == 0 {
					req.Index = 1
				}

				fmt.Println("Sending Fibonacci request: " +
					strconv.Itoa(req.Index))

				resp := fibonacci.Response{}
				err := fibserviceClient.Send(nil, "Index", req, &resp)

				if err == nil && resp.Index != 0 && resp.Value != 0 {
					fmt.Println("Fibonacci returned: " +
						strconv.FormatUint(resp.Value, 10))
					successfulRequests.Add(1)
				}
			}

		}
	}
}

func watchSignals(c chan os.Signal, quitChan chan bool) {
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSEGV,
		syscall.SIGSTOP, syscall.SIGTERM)

	for {
		select {
		case sig := <-c:
			switch sig.(syscall.Signal) {
			// Trap signals for clean shutdown
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT,
				syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM:

				quitChan <- true
				return
			}
		}
	}
}
