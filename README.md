# hw2_signer

In this task, we write an analogue unix pipeline, some like this:
```
grep 127.0.0.1 | awk '{print $2}' | sort | uniq -c | sort -nr
```

When the STDOUT of one program is passed as STDIN to another program

But in our case, these roles are performed by channels that we pass from one function to another.

The task itself essentially consists of two parts.
* Writing an ExecutePipeline function that provides us with pipeline processing of worker functions that do something.
* Writing several functions that consider us some conditional hash sum from the input data

The calculation of the hash sum is implemented by the following chain:
* SingleHash compute value crc32(data)+"~"+crc32(md5(data)) ( concat strings with ~), where data - what came to the input (in fact, the numbers from the first function)
* MultiHash compute value crc32(th+data)) (concatenation of a digit cast to a string and a string), where th=0..5 ( thus 6 hashes for each input value ), then takes the concatenation of the results in the order of calculation (0..5), where data - what came in at the entrance (and went out at the output from SingleHash)
* CombineResults gets all results, sorts (https://golang.org/pkg/sort/), concatenates sorted result with _ (underscore character) into one string
* crc32 are computed with DataSignerCrc32
* md5 are computed with DataSignerMd5

What's the catch:
* DataSignerMd5 can only be called once at a time, counts as 10ms. If several start up at the same time, there will be an overheat for 1 second
* DataSignerCrc32, counted as 1 sec
* We have 3 seconds for all calculations.
* If you do it linearly - for 7 elements it will take almost 57 seconds, so you need to somehow parallelize it

The results that are displayed if you send 2 values (commented out in the test):
```
0 SingleHash data 0
0 SingleHash md5(data) cfcd208495d565ef66e7dff9f98764da
0 SingleHash crc32(md5(data)) 502633748
0 SingleHash crc32(data) 4108050209
0 SingleHash result 4108050209~502633748
4108050209~502633748 MultiHash: crc32(th+step1)) 0 2956866606
4108050209~502633748 MultiHash: crc32(th+step1)) 1 803518384
4108050209~502633748 MultiHash: crc32(th+step1)) 2 1425683795
4108050209~502633748 MultiHash: crc32(th+step1)) 3 3407918797
4108050209~502633748 MultiHash: crc32(th+step1)) 4 2730963093
4108050209~502633748 MultiHash: crc32(th+step1)) 5 1025356555
4108050209~502633748 MultiHash result: 29568666068035183841425683795340791879727309630931025356555

1 SingleHash data 1
1 SingleHash md5(data) c4ca4238a0b923820dcc509a6f75849b
1 SingleHash crc32(md5(data)) 709660146
1 SingleHash crc32(data) 2212294583
1 SingleHash result 2212294583~709660146
2212294583~709660146 MultiHash: crc32(th+step1)) 0 495804419
2212294583~709660146 MultiHash: crc32(th+step1)) 1 2186797981
2212294583~709660146 MultiHash: crc32(th+step1)) 2 4182335870
2212294583~709660146 MultiHash: crc32(th+step1)) 3 1720967904
2212294583~709660146 MultiHash: crc32(th+step1)) 4 259286200
2212294583~709660146 MultiHash: crc32(th+step1)) 5 2427381542
2212294583~709660146 MultiHash result: 4958044192186797981418233587017209679042592862002427381542

CombineResults 29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542
```

Run as `go test -v -race`
