package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var uuid = []string{
	"bb8e38e4-dc60-11e1-8584-7f5ad6313c36",
	"bb8f130e-dc60-11e1-a7ef-0b08d979e792",
	"bb8fe036-dc60-11e1-9b6b-2707cf0708e1",
	"bb90854a-dc60-11e1-8972-87120cb186a8",
	"bb9121ee-dc60-11e1-95d1-1f0847d3bf49",
	"bb91b334-dc60-11e1-b574-df8641a0cf55",
	"bb924d30-dc60-11e1-b08d-bf05c09a4576",
	"bb92e204-dc60-11e1-bf79-bbc146575cbd",
	"bb938236-dc60-11e1-a8ba-2f541979b26b",
	"bb940f8a-dc60-11e1-87f3-33fd5d1669a4",
	"bb949b12-dc60-11e1-a00b-1f0f65398837",
	"bb95244c-dc60-11e1-9885-6beadd87979e",
	"bb95b22c-dc60-11e1-8703-1b492dc8a784",
	"bb96376a-dc60-11e1-a9a3-df3b2281df2b",
	"bb96c950-dc60-11e1-805e-bb9fc539de5b",
	"bb97503c-dc60-11e1-8432-fb52624f5fae",
	"bb97e6aa-dc60-11e1-ba12-eb936406072b",
	"bb987d40-dc60-11e1-bbdc-e320d31a9dda",
	"bb990940-dc60-11e1-b041-0b5d1484e5be",
	"bb9993ba-dc60-11e1-b3b9-dfa2db5ab272",
}

var logs []string
var logindex int

// generate pseudorandom (repeatable) logs from HHGTTG, each 
// prepended with a timestamp and with one of twenty uuids 
func logbroadcast() {
	file, err := os.Open("book.txt")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	logs = strings.Split(string(b), "\n")
	for i := range logs {
		logs[i] = time.Now().String() + "\t" + uuid[rand.Intn(len(uuid))] + "\t" + logs[i]
	}

	for ;; logindex++ {
		if logindex >= len(logs) {
			logindex = 0
		}
		h.broadcast <- logs[logindex]
		time.Sleep(time.Duration(rand.Int63n(2500)+500) * time.Millisecond)
	}
}

func dump(c chan string) {
	for i := 0; i < logindex; i++ {
		// what if logindex wraps in the middle? this won't happen in a non-dummy log (append-only)
		c<- logs[i]
	}
}


