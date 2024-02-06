package main

import (
	"log"
	"time"
)

func main() {
	channel := make(chan bool, 5)
	channel <- true
	channel <- true
	channel <- true
	channel <- true
	channel <- true

	for i := 0; i < 6; i++ {
		log.Print(<-channel)
		time.Sleep(3 * time.Second)
	}
}
