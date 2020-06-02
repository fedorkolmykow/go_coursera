package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

//сюда писать код
func ExecutePipeline(workers ...job) {
	chIn := make(chan interface{}, 0)
	chOut := make(chan interface{}, 0)
	wg := &sync.WaitGroup{}
   for _, w := range workers {
  	    wg.Add(1)

       go func(in, out chan interface{}, w job, wg *sync.WaitGroup){
		    defer wg.Done()
           defer close(out)
			w( in, out)
       }(chIn, chOut, w, wg)

  	    chIn = chOut
	    chOut = make(chan interface{}, 0)
	}
	wg.Wait()
}
//
//
func SingleHash(in, out chan interface{}){
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for val := range in {
		data := strconv.Itoa(val.(int))
		wg.Add(1)
		go SingleHashParallel(out, data, wg, mu)
	}
	wg.Wait()
}

func SingleHashParallel(out chan interface{}, data string, wg *sync.WaitGroup, mu *sync.Mutex){
	ch := make(chan string, 0)
	chmd := make(chan string, 0)

	go ParallelCrc32(data, ch)

	go func() {
		mu.Lock()
		chmd <- DataSignerMd5(data)
		mu.Unlock()
	}()

	go ParallelCrc32(<-chmd, chmd)

	out <- (<-ch) + "~" + (<-chmd)
	wg.Done()
}

func ParallelCrc32(data string, ch chan string) {
	ch <- DataSignerCrc32(data)
}

func MultiHash(in, out chan interface{}){
	wg := &sync.WaitGroup{}
	for val := range in {
		data := val.(string)
		wg.Add(1)
        go MultiHashParallel(out, data, wg)
	}
	wg.Wait()
}

func MultiHashParallel(out chan interface{}, data string, wg *sync.WaitGroup){
	result := ""
	var chans [6]chan string
	for i := 0; i < 6; i++ {
		chans[i] = make(chan string, 0)
		go ParallelCrc32(strconv.Itoa(i) + data, chans[i])
	}
	for i := 0; i < 6; i++ {
		result += <-chans[i]
	}
	out <- result
	wg.Done()
}


func CombineResults(in, out chan interface{}){
	data := []string{}

	for val := range in {
		v := val.(string)
		data = append(data, v)
	}
	sort.Strings(data)
	out <- strings.Join(data, "_")
}