BytePipe
-------------------

Needed a higher performance way of copying bytes from a UDP socket to a client parsing process.

Originally just used a chan []byte to send each read through a channel but this was slow and generated quite a bit of garbage.


Results on lenovo w540

```
go version go1.7.1 linux/amd64
BenchmarkChannel1024-8    	 5000000	       375 ns/op
BenchmarkBytePipe1024-8   	10000000	       106 ns/op
```
