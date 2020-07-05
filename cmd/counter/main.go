package main

import (
	"bufio"
	"counter/pkg/counter"
	"fmt"
	"os"
	"time"
)

func main() {
	cnt := counter.NewCounter(5, "Go", time.Minute*10)
	scanner := bufio.NewScanner(os.Stdin)
	length := 0

	for scanner.Scan() {
		txt := scanner.Text()
		fmt.Println(txt)
		cnt.RequestHTTPCount(txt)
		length++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	total := 0
	for count := range cnt.Results() {
		fmt.Printf("Count for %s: %d\n", count.URL, count.N)
		total = total + count.N
		length--

		if length == 0 {
			break
		}
	}

	fmt.Println("Total:", total)

	cnt.Stop()
}
