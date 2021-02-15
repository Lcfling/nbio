package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lesismal/nbio"
)

var (
	addrs = []string{"localhost:8888", "localhost:8888"}
)

func main() {
	var (
		wg         sync.WaitGroup
		qps        int64
		bufsize    = 64 //1024 * 8
		clientNum  = 128
		totalRead  int64
		totalWrite int64
	)

	g := nbio.NewGopher(nbio.Config{NPoller: 1})
	defer g.Stop()

	g.OnData(func(c *nbio.Conn, data []byte) {
		atomic.AddInt64(&qps, 1)
		atomic.AddInt64(&totalRead, int64(len(data)))
		atomic.AddInt64(&totalWrite, int64(len(data)))
		c.Write(append([]byte(nil), data...))
	})

	err := g.Start()
	if err != nil {
		fmt.Printf("Start failed: %v\n", err)
	}

	for i := 0; i < clientNum; i++ {
		wg.Add(1)
		idx := i
		data := make([]byte, bufsize)
		go func() {
			c, err := nbio.Dial("tcp", addrs[idx%2])
			if err != nil {
				fmt.Printf("Dial failed: %v\n", err)
			}
			g.AddConn(c)
			c.Write([]byte(data))
			atomic.AddInt64(&totalWrite, int64(len(data)))
		}()
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			fmt.Println(g.State().String())
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Printf("qps: %v, total read: %.1f M, total write: %.1f M\n", atomic.SwapInt64(&qps, 0), float64(atomic.SwapInt64(&totalRead, 0))/1024/1024, float64(atomic.SwapInt64(&totalWrite, 0))/1024/1024)
		}
	}()

	wg.Wait()
}