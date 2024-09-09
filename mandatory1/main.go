package main

import (
	"fmt"
	"math/rand"
	"time"
)

const forksNeeded = 2
const tableSize = 5 // how many philosophers and forks; to avoid magic numbers

func think(philoID int, timeInMs int) {
	fmt.Printf("Philosopher #%d thinking for %d ms\n", philoID, timeInMs)
	time.Sleep(time.Duration(timeInMs) * time.Millisecond)
}

func eat(philoID int, numTimesEaten *int) {
	*numTimesEaten++
	fmt.Printf("Philosopher #%d eating portion %d\n", philoID, *numTimesEaten)
	time.Sleep(time.Duration(300) * time.Millisecond)
}

// forks recieve requests in the form of philosopher IDs
// and second requests in the form of a bool to the relevant channel
// a fork locks untill it recieves a request, which it then processes and repeats
func fork(outChannels map[int]chan bool, inChannel chan int) {
	inUse := false
	curUser := -1
	for {
		reqID := <-inChannel

		if inUse && curUser == reqID {
			inUse = false
		} else if !inUse {
			inUse = true
			curUser = reqID
			outChannels[reqID] <- true
		} else {
			outChannels[reqID] <- false
		}
	}
}

// philosophers send their ID's to forks, which is a request to use the fork
// unless it's already in use by this philosopher, in which case it means to free it
// and recieve bools as responses for whether or not they get to use a fork
// philosophers sleep(think) out of their own volition, but other than that they only lock to wait for responses from the forks
func philosopher(ID int, outChannels []chan int, inChannels []chan bool) {

	numForksInUse := 0
	numTimesEaten := 0
	gotFork := make([]bool, len(outChannels))

	for {
		// ask for forks
		for i := 0; i < len(outChannels); i++ {
			outChannels[i] <- ID
		}

		// check how many forks we got
		for i := 0; i < len(inChannels); i++ {
			gotFork[i] = <-inChannels[i]
			if gotFork[i] {
				numForksInUse++
			}
		}

		if numForksInUse >= forksNeeded {
			eat(ID, &numTimesEaten)
		}

		// free the fork resources
		// this is part of avoiding deadlock; philosophers give up their forks whether they got to eat or not
		// thus we avoid everyone hogging 1 fork, which is deadlock
		for i := 0; i < len(gotFork); i++ {
			if gotFork[i] {
				outChannels[i] <- ID
			}
		}

		// the random think durations is another key part of avoiding deadlock; it's extremely unlikely that two philosophers think for exactly the same time
		// thus there will always be one who finishes thinking before everyone else, and gets to eat.
		// and because the think durations are random, law of large numbers ensures that the philosophers should get to eat an equal amount of times over a long period
		// and specifically, the likelihood of any philosopher eating less than 3 times drops off very rapidly
		// likewise philosophers being "nice", in always thinking after eating, ensures that everyone gets to eat
		// if they weren't "nice", you might end up with a philosopher hogging two forks and his neighbours thus never getting to eat(because he would instantly request the fork after releasing it )
		think(ID, rand.Intn(2000))
		numForksInUse = 0
	}
}

func main() {
	var forkChannels [tableSize](chan int)
	var forkPhiloMap [tableSize](map[int](chan bool))
	for i := 0; i < tableSize; i++ {
		forkPhiloMap[i] = make(map[int]chan bool)
		for k := 0; k < tableSize; k++ { // bit inefficient but problem size is small
			forkPhiloMap[i][k] = make(chan bool)
		}
		forkChannels[i] = make(chan int)
		go fork(forkPhiloMap[i], forkChannels[i])
	}

	for i := 0; i < tableSize; i++ {
		philoToForkChannels := make([](chan int), 2)
		forkToPhiloChannels := make([](chan bool), 2)
		if i == 0 { // since our heuristic is philosopher i can grab fork i-1 and i, we need special handling for i=0
			philoToForkChannels[0] = forkChannels[tableSize-1]
			philoToForkChannels[1] = forkChannels[0]
			forkToPhiloChannels[0] = forkPhiloMap[tableSize-1][i]
			forkToPhiloChannels[1] = forkPhiloMap[0][i]

		} else {
			philoToForkChannels = forkChannels[i-1 : i+1]
			forkToPhiloChannels[0] = forkPhiloMap[i-1][i]
			forkToPhiloChannels[1] = forkPhiloMap[i][i]
		}
		go philosopher(i, philoToForkChannels, forkToPhiloChannels)
	}

	for {

	}

}
