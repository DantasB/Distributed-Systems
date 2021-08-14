package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	semaphore "golang.org/x/sync/semaphore"
)

//consumer limit
var m = 100000

//memory instantiation
var memory []int

//seed of random number
var seed = rand.New(rand.NewSource(time.Now().UnixNano()))

//semaphores
var empty *semaphore.Weighted
var full *semaphore.Weighted
var mutex sync.Mutex

// generateRandomNumber receives nothing and returns a integer.
// It will use a seed that generates a number from 1 to 10^7.
// It returns the integer random number.
func generateRandomNumber() int {
	return seed.Intn(int(math.Pow(10, 7))) + 1
}

// createArrayWithZeros receives n and returns a vector of zeroes.
// It will instantiate a vector with size n.
// It returns the array containing only zeros.
func createArrayWithZeros(n int) []int {
	return make([]int, n)
}

// GetSquareRoot receives an integer number and returns this square root.
// It's necessary to cast the integer to float64 because of the sqrt function.
// Ceil the obtained square root because the output is float64.
// Convert again to integer because of the output of this function.
// It returns the square root of a number.
func getSquareRoot(number int) int {
	return int(math.Ceil(math.Sqrt(float64(number))))
}

// isPrime receives an integer number and returns a string.
// It will iterate over 2 to the square root of the number - 1.
// Check if the number is divisible by the i.
// It returns the string false or true.
func isPrime(number int) string {
	for i := 2; i < getSquareRoot(number); i++ {
		if number%i == 0 {
			return "false"
		}
	}
	return "true"
}

// getFreePosition receives nothing and returns a integer.
// It will iterate over the array.
// Check if the value of the index is 0.
// It returns -1 if there's no empty space.
// or returns the index if the array in that index is equal 0.
func getFreePosition() int {
	for i := 0; i < len(memory); i++ {
		if memory[i] == 0 {
			return i
		}
	}
	return -1
}

// getFirstFullPosition receives nothing and returns a integer.
// It will iterate over the array.
// Check if the value of the index is different from 0.
// It returns the index if the array in that index is different from 0.
// or returns -1 if all values of the array is equal 0.
func getFirstFullPosition() int {
	for i := 0; i < len(memory); i++ {
		if memory[i] != 0 {
			return i
		}
	}
	return -1
}

// consumes receives an index of a global array and fills it with 0.
func consumes() {
	var value = getFirstFullPosition()
	fmt.Printf("Is Value %d Prime? %s\n", memory[value], isPrime(memory[value]))
	memory[value] = 0
	m-- //Race condition
}

// produces receives an index of a global array and fills it with a random number.
func produces() {
	memory[getFreePosition()] = generateRandomNumber()
}

func producer() {
	ctx := context.Background()
	for {
		empty.Acquire(ctx, 1)
		mutex.Lock()
		produces()
		mutex.Unlock()
		full.Release(1)
	}

}

func consumer(finished chan bool) {
	ctx := context.Background()
	for m != 0 { //Race condition
		full.Acquire(ctx, 1)
		mutex.Lock()
		consumes() //Race condition
		mutex.Unlock()
		empty.Release(1)
	}

	finished <- true
}

func setFullToZero(n int) {
	ctx := context.Background()
	for i := 0; i < n; i++ {
		full.Acquire(ctx, 1)
	}
}

func main() {
	var np int
	var nc int
	var n int
	flag.IntVar(&np, "np", 0, "Number of Producer Threads")
	flag.IntVar(&nc, "nc", 0, "Number of Consumer Threads")
	flag.IntVar(&n, "n", 0, "Shared Memory Size")
	flag.Parse()

	if n < 1 || nc < 1 || np < 1 {
		fmt.Print("Incorrect flags values passed \n")
		return
	}

	finished := make(chan bool)
	memory = createArrayWithZeros(n)
	full = semaphore.NewWeighted(int64(n))
	setFullToZero(n)
	empty = semaphore.NewWeighted(int64(n))

	start := time.Now()
	for i := 0; i < nc; i++ {
		go consumer(finished) //race condition
	}

	for i := 0; i < nc; i++ {
		go producer()
	}

	<-finished
	duration := time.Since(start)
	fmt.Printf("Average Time Elapsed: %v seconds. For Np:%v, Nc:%v and N:%v \n", duration.Seconds(), np, nc, n)
	fmt.Print("=====================\n")
}
