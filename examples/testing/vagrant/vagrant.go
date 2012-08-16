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
var failedRequests = expvar.NewInt("failed-requests")

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	quitChan := make(chan bool, 1)
	requestChan := make(chan string, *requests*3)
	workerQuitChan := make(chan *sync.WaitGroup, *requests)

	go watchSignals(c, quitChan)

	fmt.Printf("Starting %d Workers\n", *requests)
	for i := 0; i < *requests; i++ {
		go worker(requestChan, workerQuitChan)
	}

	for {
		select {
		case <-quitChan:
			wg := new(sync.WaitGroup)

			for i := 0; i < *requests; i++ {
				wg.Add(1)
				workerQuitChan <- wg
			}

			wg.Wait()

			successful, _ := strconv.Atoi(successfulRequests.String())
			failed, _ := strconv.Atoi(failedRequests.String())
			total, _ := strconv.Atoi(totalRequests.String())

			percentSuccess := (successful / total) * 100
			percentFailed := (failed / total) * 100

			fmt.Printf("Total Requests: %d, Successful: %d (%d%%), Failed: %d (%d%%)\n", total, successful, percentSuccess, failed, percentFailed)
			return
		default:
			totalRequests.Add(1)
			requestChan <- "foo"
		}
	}
}

func worker(requestChan chan string, quitChan chan *sync.WaitGroup) {
	for {
		select {
		case wg := <-quitChan:
			wg.Done()
			break

    case _ = <-requestChan:
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
