package main

import "time"

func main() {
	ch := make(chan int)
	go func(ch chan int) {
		time.Sleep(10 * time.Second)
		// print(<-ch)
	}(ch)
	ch <- 10 /*Blocked: No routine is waiting for the data to be consumed from the channel */
}
