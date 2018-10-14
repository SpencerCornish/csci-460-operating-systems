/**************************************************************************
Spencer Cornish
CSCI460 - Operating Systems: Processor Management Assignment #1
Written in Go: https://golang.org/

To run this lab:

0: Install Homebrew
Mac only: /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"

1: Install Go
Windows: https://golang.org/dl/
Mac: brew install go

2: restart your terminal to get new path variables

3: run the lab
Both: go run path/to/lab/cornish-1.go

**************************************************************************/
package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// std_no: the last four digits of my student ID number **Renamed to stdNo to conform to language standards**
// numGeneratedJobs: the number of jobs we need to generate later
const (
	stdNo            = 0043
	numGeneratedJobs = 100
)

// A struct describing a job
type job struct {
	number         int
	arrivalTime    int
	processingTime int
}

func main() {

	// Compute the number of processors to use
	numProcessors := stdNo%3 + 2

	// Call the helper function to get user input
	algorithmChoice, runs, useCustomJobs := getUserInput()

	// A channel to recieve the results of each run
	runResultChannel := make(chan time.Duration, runs)

	for i := 0; i < runs; i++ {
		var jobs []job
		// If the user wants to use the set of 12 jobs in the assignment doc
		if useCustomJobs {
			jobs = customJobs()
		} else {
			jobs = generateJobs(numGeneratedJobs)
		}

		// 1 = circular
		// 2 = first available
		if algorithmChoice == 1 {
			go runCircular(runResultChannel, jobs, numProcessors)
		} else if algorithmChoice == 2 {
			// Start the simulation
			go runFirstAvailable(runResultChannel, jobs, numProcessors)
		} else {
			fmt.Println("Inavlid Algorithm Choice!")
			os.Exit(1)
		}
	}

	timeResultsNs := make([]int64, runs)
	minimumTimeNs := int64(math.MaxInt64)
	maximumTimeNs := int64(0)
	totalTimeNs := int64(0)

	for i := 0; i < runs; i++ {

		// Block for the next run to finish
		time := <-runResultChannel
		curTimeNs := time.Nanoseconds()
		timeResultsNs[i] = curTimeNs
		totalTimeNs = totalTimeNs + curTimeNs
		if curTimeNs < minimumTimeNs {
			minimumTimeNs = curTimeNs
		}
		if curTimeNs > maximumTimeNs {
			maximumTimeNs = curTimeNs
		}
	}

	fmt.Println("All Jobs Complete!")
	minMs := (minimumTimeNs / 1000000)
	maxMs := (maximumTimeNs / 1000000)
	meanMs := (totalTimeNs / 1000000) / int64(runs)

	// Standard Deviation
	stdDevMs := float64(0)
	n := float64(0)
	for i := 0; i < runs; i++ {
		n += math.Pow(float64((timeResultsNs[i]/1000000)-meanMs), 2)
	}
	stdDevMs = math.Sqrt(float64(n / float64(runs)))

	fmt.Printf("\nMin: %.2dms \nMax: %.2dms \nAvg: %.2dms \nStd_dev: %.2fms\n", minMs, maxMs, meanMs, stdDevMs)

	if useCustomJobs {
		fmt.Printf("Ran %d full runs, each with 12 jobs successfully\n", runs)
	} else {
		fmt.Printf("Ran %d full runs, each with %d jobs successfully\n", runs, numGeneratedJobs)
	}
}

// Run the simulation with a first Available queue. All processors pick the next job off the stack
// Returns an int representing the turnaround time in ms
func runCircular(runResultChannel chan time.Duration, jobs []job, processors int) {

	// Create the communication channels for our processor routines to use to communicate back to the main thread
	// for circular, each processor has it's own job channel
	jobChannels := make([]chan job, processors)
	for i := range jobChannels {
		jobChannels[i] = make(chan job, len(jobs))
	}

	// Result channel is how we recieve the results of each job
	jobCompleteChannel := make(chan int, len(jobs))

	// Fire up the processors
	for p := 0; p < processors; p++ {
		go processor(p, jobChannels[p], jobCompleteChannel)
	}

	startSimTime := time.Now()
	// start adding jobs, but only add a job if it's start time has occurred
	for jobIndex, job := range jobs {
		blockForArrival(startSimTime, job)

		assignedProcessorID := (jobIndex + 1) % processors
		// Put the job on the appropriate job channel
		jobChannels[assignedProcessorID] <- job
	}

	// We are done adding jobs to the jobChannels, so close them to tell the processors to terminate
	for _, jobChannel := range jobChannels {
		close(jobChannel)
	}

	// Wait for the last jobs to finish
	for i := 0; i < len(jobs); i++ {
		<-jobCompleteChannel
	}
	runDuration := time.Since(startSimTime)

	runResultChannel <- runDuration
}

// Run the simulation with a first Available queue. All processors pick the next job off the stack
// Returns an int representing the turnaround time in ms
func runFirstAvailable(runResultChannel chan time.Duration, jobs []job, processors int) {

	// Create the communication channels for our processor routines to use to communicate back to the main thread
	// jobChannel is how we give the processors their jobs.
	jobChannel := make(chan job, len(jobs))
	// Result channel is how we recieve the results of each job
	jobCompleteChannel := make(chan int, len(jobs))

	// Fire up the processors
	for p := 0; p < processors; p++ {
		go processor(p, jobChannel, jobCompleteChannel)
	}

	startSimTime := time.Now()
	// start adding jobs, but only add a job if it's start time has occurred
	for _, job := range jobs {

		blockForArrival(startSimTime, job)
		// Put the job on the jobChannel
		jobChannel <- job
	}
	// We are done adding jobs to the jobChannel, so close it to tell the processors to terminate
	close(jobChannel)

	// Wait for the jobs to finish
	for i := 0; i < len(jobs); i++ {
		<-jobCompleteChannel
	}
	runDuration := time.Since(startSimTime)

	runResultChannel <- runDuration
}

// This is a single processor, which processes jobs coming in on the jobs channel
func processor(id int, jobs <-chan job, turnaroundTime chan<- int) {
	// This is blocking for the goroutine - it will wait for a job to be assigned.
	for j := range jobs {
		// Sleep for the millisecond it takes to load in the job
		time.Sleep(time.Millisecond)
		time.Sleep(time.Duration(int64(j.processingTime * 1000000)))
		// Let the main goroutine know this job is done
		turnaroundTime <- 0
	}
}

////////////////////////////
//
// Helper Functions
//
////////////////////////////

// A helper function to get user input at the beginning of the runtime
func getUserInput() (int, int, bool) {
	reader := bufio.NewReader(os.Stdin)

	var choice int
	for {
		fmt.Print("Which Algorithm? (enter 1 for circular, 2 for firstAvailable):  ")
		algorithmChoice, _ := reader.ReadString('\n')
		choice, _ = strconv.Atoi(strings.TrimSpace(algorithmChoice))
		if choice == 1 || choice == 2 {
			break
		}
	}

	var runs int
	for {
		fmt.Print("Number of full runs: ")
		text, _ := reader.ReadString('\n')
		runs, _ = strconv.Atoi(strings.TrimSpace(text))
		if runs > 0 {
			break
		}
	}
	var useCustomJobs bool

	fmt.Print("Use the 12 preset jobs (y or n): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "y" || text == "Y" {
		useCustomJobs = true
	} else {
		useCustomJobs = false
	}

	return choice, runs, useCustomJobs
}

// A helper function to wait for the expected arrival time to begin processing
func blockForArrival(startSimTime time.Time, job job) {
	// We have to "re-bin" our arrivalTime int to int64, just to make sure we don't truncate a timestamp (for really long simulations)
	// Also, Go is big on nanoseconds rather than milliseconds, so convert our arrivalTime to nanoseconds
	arrTime := int64(job.arrivalTime) * 1000000
	curNs := time.Since(startSimTime).Nanoseconds()

	// If the next job hasn't "Arrived" yet, wait until the exact nanosecond it should before passing it to a processor
	if arrTime > curNs {
		timer := time.NewTimer(time.Duration(arrTime - curNs))
		<-timer.C
	}
}

// A helper function to make jobs with random processing time from 1-500ms
// Arrival time and number are always equal to their ID
func generateJobs(n int) []job {
	jobSlice := make([]job, n)
	for i := 0; i < n; i++ {
		jobSlice[i] = job{
			number:         i + 1,
			arrivalTime:    i + 1,
			processingTime: rand.Intn(500) + 1,
		}
	}
	return jobSlice
}

// A helper function that constructs and returns the custom jobs from the assignment description
func customJobs() []job {
	return []job{
		job{
			number:         1,
			arrivalTime:    4,
			processingTime: 9,
		},
		job{
			number:         2,
			arrivalTime:    15,
			processingTime: 2,
		},
		job{
			number:         3,
			arrivalTime:    18,
			processingTime: 16,
		},
		job{
			number:         4,
			arrivalTime:    20,
			processingTime: 3,
		},
		job{
			number:         5,
			arrivalTime:    26,
			processingTime: 29,
		},
		job{
			number:         6,
			arrivalTime:    29,
			processingTime: 198,
		},
		job{
			number:         7,
			arrivalTime:    35,
			processingTime: 7,
		},
		job{
			number:         8,
			arrivalTime:    45,
			processingTime: 170,
		},
		job{
			number:         9,
			arrivalTime:    57,
			processingTime: 180,
		},
		job{
			number:         10,
			arrivalTime:    83,
			processingTime: 178,
		},
		job{
			number:         11,
			arrivalTime:    88,
			processingTime: 73,
		},
		job{
			number:         12,
			arrivalTime:    95,
			processingTime: 8,
		},
	}
}
