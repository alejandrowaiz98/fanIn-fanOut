package main

import (
	"fmt"
	"math/rand"
)

func repeatFunc[T any, K any](done <-chan K, fn func() T) <-chan T {

	stream := make(chan T)

	go func() {
		defer close(stream)

		for {

			select {

			case <-done:
				return

			case stream <- fn():

			}

		}
	}()

	return stream

}

func main() {

	done := make(chan int)

	defer close(done)

	randomNumbFetcher := func() int { return rand.Intn(1000000) }

	for value := range repeatFunc(done, randomNumbFetcher) {

		fmt.Println(value)

	}

}
