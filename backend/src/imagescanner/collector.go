package main

import (
	"fmt"
)

func generateCollector(done chan<- bool) chan<- string {
	receiver := make(chan string, 99)

	go func() {
		for newItem := range receiver {
			fmt.Println(newItem)
		}

		// TODO: save file

		// signal we are done saving
		done <- true
	}()

	return receiver
}
