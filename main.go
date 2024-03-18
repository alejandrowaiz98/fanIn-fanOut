package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

//Benchmark with only one routine: 7.5 sec
//Benchmark with one go routine per CPUCount: 1.4026507 sec

func main() {

	start := time.Now()

	done := make(chan int)

	defer close(done)

	randomNumbFetcher := func() int { return rand.Intn(1000000) }

	randIntStream := repeatFunc(done, randomNumbFetcher)

	//For limiting our go routines based on runtime.NumCPU() result for not overstress our system
	CPUCount := runtime.NumCPU()

	//"Fan out" or splitting the primeFinder func into multiple go routines for better data processing
	primeFinderChannels := make([]<-chan int, CPUCount)

	for i := 0; i < CPUCount; i++ {

		primeFinderChannels[i] = primeFinder(done, randIntStream)

	}

	fannedInStream := fanIn(done, primeFinderChannels...)

	for value := range take(done, fannedInStream, 1000) {

		fmt.Println(value)

	}

	fmt.Printf("Se demoró : %v", time.Since(start))

}

// Function that receives a func and call it until given func is done
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

// Func for "take" n numbers from the infinite generator and pass it to other chan
func take[T any, K any](done <-chan K, stream <-chan T, n int) <-chan T {

	taken := make(chan T)

	go func() {
		defer close(taken)

		for i := 0; i < n; i++ {

			select {

			case <-done:
				return

			case taken <- <-stream:
			}

		}
	}()

	return taken

}

// Function to determinate if the given ints via randIntStream channel are or not primes
func primeFinder(done <-chan int, randIntStream <-chan int) <-chan int {

	isPrime := func(randInt int) bool {

		for i := randInt - 1; i > 1; i-- {

			if randInt%i == 0 {
				return false
			}

		}

		return true

	}

	primes := make(chan int)

	go func() {

		defer close(primes)

		for {

			select {

			case <-done:
				return

			case randomInt := <-randIntStream:
				if isPrime(randomInt) {
					primes <- randomInt
				}

			}
		}

	}()

	return primes

}

func fanIn[T any](done <-chan int, channels ...<-chan T) <-chan T {

	var wg sync.WaitGroup

	fannedInStream := make(chan T)

	transfer := func(c <-chan T) {

		defer wg.Done()

		for i := range c {
			select {
			case <-done:
				return

			case fannedInStream <- i:
			}
		}

	}

	for _, c := range channels {
		wg.Add(1)
		go transfer(c)
	}

	go func() {
		wg.Wait()
		defer close(fannedInStream)
	}()

	return fannedInStream

}
