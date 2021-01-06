// Copyright 2021 GreenYun Organization. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const defaultCount = 291601

var (
	mu1   sync.Mutex
	index int = 0
	final int

	mu2     sync.Mutex
	toFile  bool           = false
	mapping map[int]string = make(map[int]string)
)

func next() int {
	mu1.Lock()
	defer mu1.Unlock()

	index++
	if index > final {
		return -1
	}

	return index
}

func init() {
	// Ctrl-c capturing
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		mu1.Lock()
		index = final + 1
		mu1.Unlock()

		// The second time SIGTERM received
		<-c
		fmt.Fprint(os.Stderr, "\r")
		log.Fatal("user interrupted")
	}()

	// Initialize usage string
	flag.Usage = func() {
		text := fmt.Sprintf(
			`[-o FILE] [-c N] [-t N] [-v] [-p]
Options:
  -o FILE  Write to FILE insted of stdout
  -c N     Parse until N pages (default 291601)
  -t N     Start N simultaneous tasks (default %d on this system)
                [N=0 means default]
  -p       Show the percentage finished
  -pp      Show percentage, time elapsed and ETA
  -v       Verbose output`,
			runtime.NumCPU())

		fmt.Fprintln(flag.CommandLine.Output(), "Usage:", os.Args[0], text)
	}
}

func write(k int, v string) {
	mu2.Lock()
	defer mu2.Unlock()

	if toFile {
		mapping[k] = v
	} else {
		fmt.Printf("\"%s\",%d\n", v, k)
	}
}

func main() {
	var progress, progressFancy, verbose bool
	var threads int
	var filename string

	flag.StringVar(&filename, "o", "", "")
	flag.IntVar(&threads, "t", 0, "")
	flag.IntVar(&final, "c", defaultCount, "")
	flag.BoolVar(&progress, "p", false, "")
	flag.BoolVar(&progressFancy, "pp", false, "")
	flag.BoolVar(&verbose, "v", false, "")
	flag.Parse()

	if threads < 0 || final <= 0 {
		log.Fatal("wrong arguments")
	}

	if threads == 0 {
		threads = runtime.NumCPU()
	}

	if progressFancy {
		progress = true
	}

	var output *os.File
	if filename != "" {
		toFile = true

		var err error
		output, err = os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer output.Close()
	}

	// Print progress
	fin := make(chan int)
	go func() {
		for i, f, start := float64(1), float64(final), time.Now().UnixNano(); i <= f; i += 1.0 {
			<-fin

			if progress {
				nsec := time.Now().UnixNano()
				ratio := i / f

				if progressFancy { // Fancy output
					elapsed := nsec - start
					remaining := int64((f-i)/(i/float64(elapsed))) / 1e9
					fmt.Fprintf(os.Stderr, "\r%c %.2f%% done, elapsed %v, eta %-12v\r",
						`|/-\`[(nsec>>28)%4],
						ratio*100.0,
						time.Duration(elapsed/1e9)*time.Second,
						time.Duration(remaining)*time.Second)
				} else {
					fmt.Fprintf(os.Stderr, "\r%c %.2f%% done\r", `|/-\`[(nsec>>28)%4], ratio*100.0)
				}

			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(threads)

	// Main multi-task parsing function
	for i := 1; i <= threads; i++ {
		go func() {
			// Match all except letters, numbers, commas, spaces and hyphens
			p := regexp.MustCompile(`[^, \-\pL\pN]|\p{Lm}`)

			for i := next(); i > 0; i = next() {
				url := "https://www.oed.com/oed2/" + strconv.FormatInt(int64(i), 10)
				resp, err := http.Get(url)
				if err != nil {
					log.Println(err)
					continue
				}

				if resp.StatusCode != 200 {
					log.Printf("HTTP returned %d %s when getting %s\n", resp.StatusCode, resp.Status, url)
					continue
				}

				doc, err := goquery.NewDocumentFromResponse(resp)
				if err != nil {
					log.Println(err)
					continue
				}

				word := doc.Find(".hwLabel").Nodes[0].FirstChild.Data
				word = p.ReplaceAllString(word, "")
				word = strings.TrimSpace(word)
				word = strings.TrimSuffix(word, ",")
				write(i, word)

				if verbose {
					log.Printf("parsed: %d\n", i)
				}
				fin <- i
			}

			wg.Done()
		}()
	}

	if progress || verbose {
		log.Printf("started %d tasks for indexing\n", threads)
	}

	wg.Wait()

	if toFile {
		if progress || verbose {
			log.Printf("writing to %s", filename)
		}

		output.WriteString("Word,Index\n")
		for k, v := range mapping {
			output.WriteString(fmt.Sprintf("\"%s\",%d\n", v, k))
		}
	}

	if progress || verbose {
		log.Println("exiting")
	}
}
