BytePipe
-------------------

Needed a higher performance way of copying bytes from a UDP socket to a client parsing process.

Originally just used a chan []byte to send each read through a channel but this was slow and generated quite a bit of garbage.

BytePipe is a blocking ringbuffer (it won't overwrite unread data)

If len==0 a read will block until some data is available.

If len==capacity a write will block until it can write everything.


Results on lenovo w540

```
go version go1.7.1 linux/amd64
BenchmarkChannel128-8     	 1000000	      1206 ns/op
BenchmarkChannel1024-8    	  500000	      3407 ns/op
BenchmarkChannel2048-8    	  300000	      4490 ns/op
BenchmarkBytePipe128-8    	 3000000	       436 ns/op
BenchmarkBytePipe1024-8   	 2000000	       642 ns/op
BenchmarkBytePipe2048-8   	 3000000	       860 ns/op
```
