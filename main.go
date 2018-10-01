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
Both: go run path/to/lab/main.go

**************************************************************************/
package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// std_no: the last four digits of my student ID number **Renamed to stdNo to conform to language standards**
// numGeneratedJobs: the number of jobs we need to generate later
const (
	stdNo            = 0043
	numGeneratedJobs = 100
)

// A logger reference, not yet populated
var logger log.Logger

// A struct describing a job
type job struct {
	number         int
	arrivalTime    int
	processingTime int
}

func main() {
	// Make a logger, so we can have timestamps on our prints
	logger = *log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	logger.Println("Starting simulation")

	// Compute the number of processors to use
	numProcessors := stdNo%3 + 2

	logger.Printf("We will be using %d processors running %d jobs for this simulation\n", numProcessors, numGeneratedJobs)

	// Generate a slice (AKA list) of jobs
	jobs := generateJobs(numGeneratedJobs)

	// Record a start time for the simulation
	start := time.Now()

	// Start the simulation
	totalTurnaroundTime := runSpencer(jobs, numProcessors)

	logger.Printf("TOTAL CPU TIME TIME FOR THIS RUN: %s\n", totalTurnaroundTime.String())
	logger.Printf("TOTAL TURNAROUND TIME TIME FOR THIS RUN: %s\n", time.Since(start).String())
}

// Run the simulation with a first Available queue. All processors pick the next job off the stack
// Returns an int representing the turnaround time in ms
func runSpencer(jobs []job, processors int) time.Duration {
	// A timestamp for the beginning of time, used to add turnaround durations to
	// TODO: Figure out if we need to do this at all
	totalTurnaround := time.Time{}

	// Create the communication channels for our processor routines to use to communicate back to the main thread
	// jobChannel is how we give the processors their jobs.
	jobChannel := make(chan job, len(jobs))
	// Result channel is how we recieve the results of each job
	resultChannel := make(chan time.Duration, len(jobs))

	// Fire up the processors
	for p := 0; p < processors; p++ {
		go processor(p, jobChannel, resultChannel)
	}

	startSimTime := time.Now()
	// start adding jobs, but only add a job if it's start time has occurred
	for _, job := range jobs {
		// We have to "re-bin" our arrivalTime int to int64, just to make sure we don't truncate a timestamp (for really long simulations)
		// Also, Go is big on nanoseconds rather than milliseconds, so convert our arrivalTime to nanoseconds
		arrTime := int64(job.arrivalTime) * 1000000
		curNs := time.Since(startSimTime).Nanoseconds()

		// If the next job hasn't "Arrived" yet, wait until the exact nanosecond it should before passing it to a processor
		if arrTime > curNs {
			timer := time.NewTimer(time.Duration(arrTime - curNs))
			<-timer.C
		}
		// Put the job on the jobChannel
		jobChannel <- job
	}
	// We are done adding jobs to the jobChannel, so close it to tell the processors to terminate
	close(jobChannel)

	// Gather up the results of the jobs
	for i := 0; i < len(jobs); i++ {
		jobDur := <-resultChannel
		totalTurnaround = totalTurnaround.Add(jobDur)
	}

	// Now that we have added the durations of each job to the zero time, subtract zero to get a duration, rather than a timeStamp
	return totalTurnaround.Sub(time.Time{})

}

func processor(id int, jobs <-chan job, turnaroundTime chan<- time.Duration) {
	// This is blocking for the goroutine - it will wait for a job to be assigned.
	for j := range jobs {
		// store off the start time
		start := time.Now()
		// Sleep for the millisecond it takes to load in the job
		time.Sleep(time.Millisecond)
		logger.Printf("Job Loaded on processor ID #%d\n", id)
		durationToSleep, _ := time.ParseDuration(strconv.Itoa(j.processingTime) + "ms")
		time.Sleep(durationToSleep)
		turnaroundTime <- time.Since(start)
		logger.Printf("Job Completed on processor ID #%d in %s \n", id, time.Since(start).String())
	}
}

// A Helper function to choose a processor to send a job to
func chooseProcessor(jobNumber, numProcessors int) int {
	return (jobNumber + 1) % numProcessors
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
