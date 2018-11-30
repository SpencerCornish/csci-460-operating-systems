/**************************************************************************
Spencer Cornish
CSCI460 - Operating Systems: Priority Inversion - Assignment 3
Written in Go: https://golang.org/

To run this lab:

0: Install Homebrew
Mac only: /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"

1: Install Go
Windows: https://golang.org/dl/
Mac: brew install go

2: restart your terminal to get new path variables

3: run the lab
Both: go run path/to/lab/cornish-3.go

**************************************************************************/
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type job struct {
	// A unique id for the job (incremented from zero on generation/load)
	uid int

	// The time the task arrived
	arrivalTime int

	// The type of job. An int 1-3
	jobType int

	// The number of computational slices this job has recieved.
	completedTicks int
}

type buffer struct {
	bufferedString bytes.Buffer
	lockedBy       int
}

// Used to get subscript chars for later
var subscript = []rune("₀₁₂₃₄₅₆₇₈₉")

func main() {

	// Get user choice on what kind of jobs to run
	jobsToGenerate := getUserInput()

	// Generate and sort the jobs
	jobs := generateJobs(jobsToGenerate)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].arrivalTime < jobs[j].arrivalTime
	})

	// Track our current place in time, starting at tick 1
	tick := 1

	// Buffers for the three job types
	oneAndThreeBuffer := buffer{}
	twoBuffer := buffer{}

	// Pop the first job
	currentJob := jobs[0]
	jobs = jobs[1:]

	// The time the current job started at. Recorded at all the places a job can start
	startTime := -1

	for {
		// Tick until the current job has arrived
		if tick >= currentJob.arrivalTime {
			jobDuration := getJobDuration(currentJob.jobType)

			// lock the buffer when we start, amnd record the start time
			if (currentJob.jobType == 1 || currentJob.jobType == 3) && oneAndThreeBuffer.bufferedString.Len() == 0 {
				oneAndThreeBuffer.lockedBy = currentJob.jobType
				startTime = tick
			}
			if currentJob.jobType == 2 && twoBuffer.bufferedString.Len() == 0 {
				twoBuffer.lockedBy = currentJob.jobType
				startTime = tick
			}

			// Check all arrived jobs for a higher priority job that can preempt the current one
			for i, job := range jobs {

				// Don't check jobs that haven't arrived yet
				if job.arrivalTime > tick {
					break
				}

				// If the job has arrived, and it's priority is higher than the current job, preempt
				if tick >= job.arrivalTime && job.jobType < currentJob.jobType {

					// If the job that can preempt is locked out of it's mutex, bail
					if oneAndThreeBuffer.lockedBy != job.jobType && (job.jobType == 1 || job.jobType == 3) {
						continue
					}
					// Print out the job, since we are going to replace it with the higher priority job
					printJob(startTime, currentJob, oneAndThreeBuffer, twoBuffer)

					// grab the current job, so we can add it back into the queue after we preempt it
					oldCurrentJob := currentJob

					// Set the new current job, when it started, and the new job duration
					currentJob = job
					startTime = tick
					jobDuration = getJobDuration(currentJob.jobType)

					// Remove the job we are switching to from the queue
					jobs = append(jobs[:i], jobs[i+1:]...)

					// Readd the job that was interrupted to the queue, and resort the slice, just to be safe
					jobs = append(jobs, oldCurrentJob)
					sort.Slice(jobs, func(i, j int) bool {
						return jobs[i].arrivalTime < jobs[j].arrivalTime
					})
					break
				}

			}
			// END PREEMPTION CHECKS

			// Add to the buffer for this tick
			if currentJob.jobType == 1 || currentJob.jobType == 3 {
				oneAndThreeBuffer.bufferedString.WriteString(strconv.Itoa(currentJob.jobType))
			} else if currentJob.jobType == 2 {
				twoBuffer.bufferedString.WriteString("N")
			}
			currentJob.completedTicks++

			// If the job is done, move to the next job, or break if all jobs complete
			if currentJob.completedTicks == jobDuration {

				// Print the job and reset the mutex/buffer
				printJob(startTime, currentJob, oneAndThreeBuffer, twoBuffer)
				if currentJob.jobType == 1 || currentJob.jobType == 3 {
					oneAndThreeBuffer.bufferedString.Reset()
					oneAndThreeBuffer.lockedBy = -1

				} else if currentJob.jobType == 2 {
					twoBuffer.bufferedString.Reset()
					twoBuffer.lockedBy = -1
				}

				// We are done if there are no more jobs to do
				if allJobsComplete(jobs) {
					break
				}
				// Grab the job that arrived first, then check any other
				// jobs that have arrived to see if they are higher priority
				newCurrentJob := jobs[0]
				newCurrentJobIdx := 0

				// This loop picks the highest priority job, that has arrived, that can access the buffer it needs
				for i, job := range jobs {
					if (tick+1) >= job.arrivalTime && job.jobType < newCurrentJob.jobType {

						// Buffer check
						if (job.jobType == 1 || job.jobType == 3) && oneAndThreeBuffer.lockedBy != job.jobType && oneAndThreeBuffer.lockedBy != -1 {
							continue
						}
						newCurrentJob = job
						newCurrentJobIdx = i
					}

				}

				currentJob = newCurrentJob
				startTime = tick + 1
				// Remove the job we are switching to from the queue
				jobs = append(jobs[:newCurrentJobIdx], jobs[newCurrentJobIdx+1:]...)
			}

		}
		// Move to the next tick
		tick++
	}

}

// Prints out the job, and other useful information about the job
func printJob(startTime int, job job, oneAndThreeBuffer buffer, twoBuffer buffer) {
	// Get the subscript representation of the jobType
	subscriptJobType := string(subscript[job.jobType])

	// Get the buffer contents for the job from the right buffer
	var printString string
	if job.jobType == 1 || job.jobType == 3 {
		printString = oneAndThreeBuffer.bufferedString.String()
	} else if job.jobType == 2 {
		printString = twoBuffer.bufferedString.String()
	}

	fmt.Printf("time %d, t%s%st%s  [job UID: %d, arrival time: %d]\n", startTime, subscriptJobType, printString, subscriptJobType, job.uid, job.arrivalTime)

}

// Calculate the duration of this job, based on the assignment constraints
func getJobDuration(prio int) int {
	if prio == 1 {
		return 3
	}
	if prio == 2 {
		return 10
	}
	if prio == 3 {
		return 3
	}
	return -1
}

// Check if all jobs are complete
func allJobsComplete(jobs []job) bool {
	for _, job := range jobs {
		if job.completedTicks != getJobDuration(job.jobType) {
			return false
		}
	}
	return true
}

func generateJobs(amount int) []job {
	// Return the special list if amount == -1
	if amount == -1 {
		return []job{
			job{
				uid:         0,
				arrivalTime: 1,
				jobType:     3,
			},
			job{
				uid:         1,
				arrivalTime: 3,
				jobType:     2,
			},
			job{
				uid:         2,
				arrivalTime: 6,
				jobType:     3,
			},
			job{
				uid:         3,
				arrivalTime: 8,
				jobType:     1,
			},
			job{
				uid:         4,
				arrivalTime: 10,
				jobType:     2,
			},
			job{
				uid:         5,
				arrivalTime: 12,
				jobType:     3,
			},
			job{
				uid:         6,
				arrivalTime: 26,
				jobType:     1,
			},
		}

	}
	generatedJobSlice := make([]job, amount)

	// Seed the random number generator
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < amount; i++ {
		generatedJobSlice[i] = job{
			uid: i,
			// Times from 1 - 30
			arrivalTime: rnd.Intn(29) + 1,
			// Priorities from 1-3
			jobType: rnd.Intn(3) + 1,
		}
	}

	return generatedJobSlice
}

// Get user input for what kind of jobs to run
func getUserInput() int {
	reader := bufio.NewReader(os.Stdin)

	var choice int
	for {
		fmt.Print("Generate Random Jobs? (enter 1 for yes, 2 for preset list):  ")
		algorithmChoice, _ := reader.ReadString('\n')
		choice, _ = strconv.Atoi(strings.TrimSpace(algorithmChoice))
		if choice == 1 || choice == 2 {
			break
		}
	}

	if choice == 2 {
		// Returning -1 will indicate to the generator that we should use the preset list of jobs
		return -1
	}

	for {
		fmt.Print("how many random jobs should be run? (Enter number > 0 and <= 10):  ")
		algorithmChoice, _ := reader.ReadString('\n')
		count, err := strconv.Atoi(strings.TrimSpace(algorithmChoice))
		if err == nil && count > 0 && count < 11 {
			return count
		}
	}

}
