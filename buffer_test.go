package bytepipe

import (
	"log"
	"testing"
)

func TestBytePipe(t *testing.T) {
	pipe := NewBytePipe(20)
	go func() {
		pipe.Write(make([]byte, 5))
		pipe.Write(make([]byte, 6))
		pipe.Write(make([]byte, 7))
		pipe.Write(make([]byte, 8))
	}()

	buf := make([]byte, 10)
	n := pipe.Read(buf)
	n2 := pipe.Read(buf)
	if n+n2 == 15 {
		// This means that we read all bytes in two reads.
		return
	}
	n = pipe.Read(buf)
	// This means some writes were still occuring when we
	if n == 0 {
		//
		t.FailNow()
	}
}

func TestBytePipeLargeMessage(t *testing.T) {
	numBytes := 50
	var pipeCap uint32 = 10
	buffer := 20

	pipe := NewBytePipe(pipeCap)
	go func(nb int) {
		// Write in them bytes!
		pipe.Write(make([]byte, nb))
	}(numBytes)

	buf := make([]byte, buffer)
	total := 0
	for total < numBytes {
		total += pipe.Read(buf)
	}
	log.Printf("Read %d bytes using a %d buffer", total, buffer)
}

// func TestOutput(t *testing.T) {
// 	pipe := NewBytePipe(0)
// 	buf := make([]byte, 1024)
// 	inbuf := make([]byte, 1024)
// 	total := 0
// 	go func() {
// 		for {
// 			n := pipe.Write(inbuf)
// 			if n == 0 {
// 				break
// 			}
// 		}
// 	}()
// 	go func() {
// 		for {
// 			n := pipe.Read(buf)
// 			if n == 0 {
// 				break
// 			}
// 			total += n
// 		}
// 	}()
//
// 	time.Sleep(time.Second * 5)
// 	fmt.Printf("BytePipe mb/sec: %.2f\n", float64(total/10)/1024.0/1024.0)
//
// 	chanPipe := make(chan []byte, 1024)
// 	total = 0
// 	go func() {
// 		for {
// 			inbuf := make([]byte, len(buf))
// 			copy(inbuf, buf)
// 			chanPipe <- inbuf
// 		}
// 	}()
// 	go func() {
// 		for {
// 			b := <-chanPipe
// 			total += len(b)
// 		}
// 	}()
//
// 	time.Sleep(time.Second * 5)
// 	fmt.Printf("Channel mb/sec: %.2f\n", float64(total/10)/1024.0/1024.0)
// }

func BenchmarkChannel128(b *testing.B) {
	pipe := make(chan []byte, 128)
	buf := make([]byte, 128)
	b.ResetTimer()
	go func(n int) {
		for i := 0; i < n; i++ {
			inbuf := make([]byte, len(buf))
			copy(inbuf, buf)
			pipe <- inbuf
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		<-pipe
	}
	b.StopTimer()
}

func BenchmarkChannel1024(b *testing.B) {
	pipe := make(chan []byte, 1024)
	buf := make([]byte, 1024)
	b.ResetTimer()
	go func(n int) {
		for i := 0; i < n; i++ {
			inbuf := make([]byte, len(buf))
			copy(inbuf, buf)
			pipe <- inbuf
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		<-pipe
	}
	b.StopTimer()
}

func BenchmarkChannel2048(b *testing.B) {
	pipe := make(chan []byte, 2048)
	buf := make([]byte, 2048)
	b.ResetTimer()
	go func(n int) {
		for i := 0; i < n; i++ {
			inbuf := make([]byte, len(buf))
			copy(inbuf, buf)
			pipe <- inbuf
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		<-pipe
	}
	b.StopTimer()
}

func BenchmarkBytePipe128(b *testing.B) {
	pipe := NewBytePipe(0)
	buf := make([]byte, 128)
	inbuf := make([]byte, 128)
	b.ResetTimer()
	go func(n int) {
		for i := 0; i < n; i++ {
			if pipe.Write(inbuf) == 0 {
				break
			}
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		if pipe.Read(buf) == 0 {
			break
		}
	}
	b.StopTimer()
}

func BenchmarkBytePipe1024(b *testing.B) {
	pipe := NewBytePipe(0)
	buf := make([]byte, 1024)
	inbuf := make([]byte, 1024)
	b.ResetTimer()
	go func(n int) {
		for i := 0; i < b.N; i++ {
			pipe.Write(inbuf)
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		pipe.Read(buf)
	}
	b.StopTimer()
}

func BenchmarkBytePipe2048(b *testing.B) {
	pipe := NewBytePipe(0)
	buf := make([]byte, 2048)
	inbuf := make([]byte, 2048)
	b.ResetTimer()
	go func(n int) {
		for i := 0; i < n; i++ {
			if pipe.Write(inbuf) == 0 {
				break
			}
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		if pipe.Read(buf) == 0 {
			break
		}
	}
	b.StopTimer()
}
