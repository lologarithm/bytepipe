package bytepipe

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestBytePipe(t *testing.T) {
	pipe := NewBytePipe(20)
	go func() {
		wtot := 0
		wtot += pipe.Write([]byte{1, 2, 3, 4, 5})
		wtot += pipe.Write([]byte{6, 7, 8, 9, 10, 11})
		wtot += pipe.Write([]byte{12, 13, 14, 15, 16, 17, 18})
		wtot += pipe.Write([]byte{19, 20, 21, 22, 23, 24, 25, 26})
		log.Printf("TestBytePipe Wrote: %d bytes\n", wtot)
		if wtot != 26 {
			t.FailNow()
		}
	}()
	time.Sleep(time.Second)
	buf := make([]byte, 15)
	n := pipe.Read(buf)
	counter := 0
	fmt.Printf("Read %d bytes\n", n)
	for i := 0; i < n; i++ {
		if int(buf[i]) != (i + counter + 1) {
			log.Printf("Byte %d value %d value was incorrect, expected: %d", i+counter, buf[i], i+counter+1)
		}
	}
	counter = n
	n2 := pipe.Read(buf)
	fmt.Printf("Read %d bytes\n", n2)
	for i := 0; i < n2; i++ {
		if int(buf[i]) != (i + counter + 1) {
			log.Printf("Byte %d value %d value was incorrect, expected: %d", i+counter, buf[i], i+counter+1)
		}
	}
	counter += n2
	if n+n2 == 26 {
		// This means that we read all bytes in two reads.
		return
	}
	n = pipe.Read(buf)
	// This means some writes were still occuring when we were reading
	if n == 0 {
		//
		t.FailNow()
	}
	for i := 0; i < n; i++ {
		if int(buf[i]) != (i + counter + 1) {
			log.Printf("Byte %d value %d value was incorrect, expected: %d", i+counter, buf[i], i+counter+1)
		}
	}
}

func TestBytePipeLargeMessage(t *testing.T) {
	// Write more bytes to buffer than fit
	// Try reading while write is still occuring
	numBytes := 50
	var pipeCap uint32 = 20
	buffer := 10

	pipe := NewBytePipe(pipeCap)
	go func(nb int) {
		// Write in them bytes!
		a := make([]byte, nb)
		for i := 0; i < nb; i++ {
			a[i] = byte(i + 1)
		}
		wrote := pipe.Write(a)
		log.Printf("TestBytePipeLarge Wrote num bytes: %d\n", wrote)
		if wrote != nb {
			t.FailNow()
		}
	}(numBytes)

	time.Sleep(time.Second)
	buf := make([]byte, buffer)
	total := 0
	for total < numBytes {
		n := pipe.Read(buf)
		total += n
		log.Printf("Read %d bytes, total: %d\n", n, total)
	}

	log.Printf("Read %d bytes using a %d buffer", total, buffer)
	if pipe.Len() != 0 {
		log.Printf("Bytes left after reading everything: %d", pipe.Len())
		t.FailNow()
	}
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
