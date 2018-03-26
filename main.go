package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strconv"
	"regexp"
	"flag"
	"os"
)

const baseAdress = "https://bugs.archlinux.org/task/"

func parse(id int) string {
	resp, err := http.Get(baseAdress + strconv.Itoa(id))
	if err != nil {
		fmt.Errorf("networking error: %s", err)
		return ""
	}

	t := html.NewTokenizer(resp.Body)
	Parser:
	for {
		tk := t.Next()

		switch tk {
		case html.ErrorToken:
			break Parser
		case html.StartTagToken:
			tz := t.Token()
			if tz.Data == "h2" {
				txt := t.Next()
				if txt == html.TextToken {
					return string(t.Text()[3:])
				}
			}
		}
	}
	return ""
}

func worker(id int, jobs <-chan int, results chan<- string) {
	for job := range jobs {
		rs := parse(job)
		results <- rs
	}
}

func main() {
	startingTask := flag.Int("start", -1, "ID of task to start from (REQUIRED)")
	count := flag.Int("count", -1, "Count of tasks to go through (REQUIRED)")
	workers := flag.Int("workers", 50, "Number of concurrent tasks to be run")
	flag.Parse()

	if *startingTask == -1 || *count == -1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	rgx := regexp.MustCompile(`FS#\d+ - (.+)`)
	failed := 0

	jobs := make(chan int, *count)
	results := make(chan string, *count)

	for w := 1; w <= *workers; w++ {
		go worker(w, jobs, results)
	}

	for i := *startingTask; i > *startingTask-*count; i-- {
		jobs <- i
	}
	close(jobs)


	for r := 1; r <= *count; r++ {
		res := <-results
		matches := rgx.FindStringSubmatch(html.UnescapeString(res))
		if len(matches) < 2 {
			failed++
			continue
		}
		fmt.Println(matches[1])
	}
}
