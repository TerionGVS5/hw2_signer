package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func _crc32(input string, resultChan chan orderedData, currIndex int) {
	result := DataSignerCrc32(input)
	resultChan <- orderedData{
		value: result,
		index: currIndex,
	}
}

func _crc32Th(input string, resultChan chan orderedDataWithTh, currIndex int, currTh int) {
	result := DataSignerCrc32(input)
	resultChan <- orderedDataWithTh{
		orderedData: orderedData{
			value: result,
			index: currIndex,
		},
		th: currTh,
	}
}

func _md5(input string, resultChan chan orderedData, currIndex int, quotaCh chan struct{}) {
	quotaCh <- struct{}{}
	result := DataSignerMd5(input)
	resultChan <- orderedData{
		value: result,
		index: currIndex,
	}
	<-quotaCh
}

type orderedData struct {
	value string
	index int
}

type orderedDataWithTh struct {
	orderedData orderedData
	th          int
}

func SingleHash(in, out chan interface{}) {
	chanCrc32data := make(chan orderedData, MaxInputDataLen)
	chanMd5data := make(chan orderedData, MaxInputDataLen)
	chanCrc32Md5data := make(chan orderedData, MaxInputDataLen)
	mapCrc32data := make(map[int]string)
	mapCrc32Md5data := make(map[int]string)
	quotaCh := make(chan struct{}, 1)
	countNumbers := 0
	for val := range in {
		input := strconv.Itoa(val.(int))
		countNumbers += 1
		go _crc32(input, chanCrc32data, countNumbers)
		go _md5(input, chanMd5data, countNumbers, quotaCh)
	}
	currCountNumbers := countNumbers
	for val := range chanMd5data {
		go _crc32(val.value, chanCrc32Md5data, val.index)
		currCountNumbers -= 1
		if currCountNumbers == 0 {
			break
		}
	}
	currCountNumbers = countNumbers
	for val := range chanCrc32Md5data {
		currCrc32 := <-chanCrc32data
		mapCrc32data[currCrc32.index] = currCrc32.value
		mapCrc32Md5data[val.index] = val.value
		currCountNumbers -= 1
		if currCountNumbers == 0 {
			break
		}
	}
	for i := 1; i <= countNumbers; i++ {
		out <- fmt.Sprintf("%s~%s", mapCrc32data[i], mapCrc32Md5data[i])
	}
}

func MultiHash(in, out chan interface{}) {
	chanCrc32dataWithTh := make(chan orderedDataWithTh, MaxInputDataLen*6)
	mapCrc32dataTh := make(map[string]string)
	countNumbers := 0
	for val := range in {
		countNumbers += 1
		for th := 0; th < 6; th++ {
			go _crc32Th(strconv.Itoa(th)+val.(string), chanCrc32dataWithTh, countNumbers, th)
		}
	}
	currCountNumbersMulTh := countNumbers * 6
	for val := range chanCrc32dataWithTh {
		mapCrc32dataTh[fmt.Sprintf("%d_%d", val.orderedData.index, val.th)] = val.orderedData.value
		currCountNumbersMulTh -= 1
		if currCountNumbersMulTh == 0 {
			break
		}
	}
	for currNumberIndex := 1; currNumberIndex <= countNumbers; currNumberIndex++ {
		currResult := ""
		for currTh := 0; currTh < 6; currTh++ {
			currResult += mapCrc32dataTh[fmt.Sprintf("%d_%d", currNumberIndex, currTh)]
		}
		out <- currResult
	}
}

func CombineResults(in, out chan interface{}) {
	results := make([]string, 0, MaxInputDataLen)
	for val := range in {
		results = append(results, val.(string))
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}

func pipeline(currJob job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	currJob(in, out)
}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	var wg sync.WaitGroup
	for index, currJob := range jobs {
		wg.Add(1)
		if index%2 == 0 {
			go pipeline(currJob, in, out, &wg)
			in = make(chan interface{}, 0)
		} else {
			go pipeline(currJob, out, in, &wg)
			out = make(chan interface{}, 0)
		}
	}
	wg.Wait()
}
