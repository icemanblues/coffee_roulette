package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type History map[string]map[string]time.Time

func oddError(n int) error {
	return fmt.Errorf("Must have an even number of people. You have %d", n)
}

// Blank represents the empty person. Used to handle odd numbered headcount in matching
const Blank string = ""

// ErrNoSolution error when there is no solution given the constraints
var ErrNoSolution error = fmt.Errorf("No Solution Possible")

// ReadHistory reads the history from a file on the filesystem
func ReadHistory(filename string) (History, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var history History
	yaml.Unmarshal(bytes, &history)
	return history, nil
}

// WriteHistory writes the history to a file on the filesystem
func WriteHistory(filename string, history History) error {
	bytes, err := yaml.Marshal(history)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Match will pair individuals from so that they are not paired with them selves
// and not with someone that they have been paired with previously
func Match(people []string, history History, result map[string]string) (map[string]string, error) {
	// do we have a valid solution?
	if len(people) == len(result) {
		return result, nil
	}

	// handle odd use case
	if l := len(people) % 2; l != 0 {
		return nil, oddError(l)
	}

	for _, p := range people {
		// was p already matched in this possible solution?
		if _, ok := result[p]; ok {
			continue
		}

		for _, q := range people {
			// was q already matched?
			if _, ok := result[q]; ok {
				continue
			}

			// can't match with yourself
			if p == q {
				continue
			}

			// were they matched previously?
			if pHist, ok := history[p]; ok {
				if _, ok := pHist[q]; ok {
					continue
				}
			}

			// try p and q  as a possibly
			result[p] = q
			result[q] = p
			sol, err := Match(people, history, result)
			if err == nil {
				return sol, nil
			}
			delete(result, p)
			delete(result, q)
		}

		// no q in people for p in current result + history
		return nil, ErrNoSolution
	}

	// we tried everything
	return nil, ErrNoSolution
}

// QuickMatch will quickly pair everyone with everyone else, then rotate by one and pair again.
// When we arrive where we started, then we are done
func QuickMatch(people []string) ([]map[string]string, error) {
	if l := len(people); l%2 != 0 {
		return nil, oddError(l)
	}

	matches := make([]map[string]string, 0, len(people))
	for n := 1; n < len(people); n++ {
		pair := map[string]string{}
		for i, p := range people {
			j := (i + n) % len(people)
			q := people[j]

			if _, ok := pair[p]; ok {
				continue
			}
			if _, ok := pair[q]; ok {
				continue
			}

			pair[p] = q
			pair[q] = p
		}
		matches = append(matches, pair)
	}
	return matches, nil
}

// AddToHistory given the existing history and a the results from Match, it will combine them into a new history
func AddToHistory(history History, result map[string]string, now time.Time) History {
	for key, val := range result {
		keyHist, ok := history[key]
		if !ok {
			keyHist = make(map[string]time.Time)
			history[key] = keyHist
		}
		keyHist[val] = now
	}

	return history
}

func main() {
	fmt.Println("coffee roulette!")
	peopleFilename := flag.String("people", "", "the people file")
	histFilename := flag.String("history", "", "the history file")
	flag.Parse()

	if *peopleFilename == "" {
		// panic("you must supply a people file")
		flag.Usage()
		return
	}

	if *histFilename == "" {
		// panic("you must supply a history file")
		flag.Usage()
		return
	}

	people := []string{"a", "b", "c", "d"}
	history, err := ReadHistory(*histFilename)
	if err != nil {
		panic(err)
	}

	result := make(map[string]string)
	answer, err := Match(people, history, result)
	if err != nil {
		fmt.Println("Unable to solve")
	}

	history = AddToHistory(history, answer, time.Now())
	WriteHistory("a.out.yml", history)
}
