package counter

import (
	"bytes"
	"container/list"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Counter struct {
	sem     chan struct{}
	k       int
	query   string
	res     chan Count
	waiters *list.List
	done    chan struct{}
	mu      sync.RWMutex
	timeout time.Duration
}

func NewCounter(k int, query string, timeout time.Duration) *Counter {
	if k < 1 {
		k = 1
	}

	cnt := &Counter{k: k,
		sem:     make(chan struct{}, k),
		query:   query,
		res:     make(chan Count),
		waiters: list.New(),
		done:    make(chan struct{}, 1),
		timeout: timeout,
	}

	go cnt.handle()

	return cnt
}

func (c *Counter) RequestHTTPCount(url string) {
	c.mu.Lock()
	c.waiters.PushBack(func() {
		n, _ := fetch(url, c.query)
		timeout := time.After(c.timeout)
	cl:
		for {
			select {
			case c.res <- Count{URL: url, N: n}:
				break cl
			case <-timeout:
				break cl
			}
		}
		<-c.sem
	})
	c.mu.Unlock()
}

func (c *Counter) handle() {
	for {
		select {
		case c.sem <- struct{}{}:
			c.mu.Lock()
			if e := c.waiters.Front(); e != nil {
				go c.waiters.Remove(e).(func())()
			} else {
				<-c.sem
			}
			c.mu.Unlock()
		case <-c.done:
			return
		}
	}
}

func (c *Counter) Results() <-chan Count {
	return c.res
}

func (c *Counter) Stop() {
	c.done <- struct{}{}
	close(c.res)
}

func fetch(url string, query string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return 0, err
	}

	return bytes.Count(data, []byte(query)), nil
}

type Count struct {
	URL string
	N   int
}
