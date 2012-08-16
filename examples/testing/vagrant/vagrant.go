package main

import (
	"expvar"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

var requests = flag.Int("requests", 10, "number of concurrent requests")
var doozer = flag.String("doozer", "127.0.0.1:8046", "doozer instance to connect to")

var totalRequests = expvar.NewInt("total-requests")
var successfulRequests = expvar.NewInt("successful-requests")

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	quitChan := make(chan bool, 1)
	requestChan := make(chan string, *requests*3)
	workerQuitChan := make(chan bool, 1)
  workerWaitGroup := new(sync.WaitGroup)

	go watchSignals(c, quitChan)

	fmt.Printf("Starting %d Workers\n", *requests)
	for i := 0; i < *requests; i++ {
		go worker(requestChan, workerWaitGroup, workerQuitChan)
	}

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
			requestChan <- "foo"
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

		case _ = <-requestChan:
			totalRequests.Add(1)
      fmt.Println("sending request")
			successfulRequests.Add(1)
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
			}
		}
	}
}
