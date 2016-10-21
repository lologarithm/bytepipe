package bytepipe

import (
	"sync"
	"sync/atomic"
)

// BytePipe is a high performance single reader/single writer
// threadsafe pipe to shove bytes through.
type BytePipe struct {
	wbuf []byte // Internal buffer
	wi   uint32 // Write index

	s       uint32 // s==state 1==alive 0==closed
	cap     uint32 // capacity of the buffer
	written uint32 // Bytes written

	rbuf []byte // Internal buffer
	ri   uint32 // read index

	wrCond *sync.Cond
}

// NewBytePipe creates a new byte pipe with max buffer size.
// Defaults to 32768 bytes if cap == 0
func NewBytePipe(cap uint32) *BytePipe {
	if cap == 0 {
		cap = 32768 // 32kb of room for buffering.
	}
	buf := make([]byte, cap)
	return &BytePipe{
		rbuf:    buf,
		wbuf:    buf,
		s:       1,
		written: 0,
		wi:      0,
		ri:      0,
		cap:     cap,
		wrCond:  sync.NewCond(&sync.Mutex{}),
	}
}

// Len returns the number of unread bytes in the buffer.
func (bp *BytePipe) Len() int {
	return int(atomic.LoadUint32(&bp.written))
}

// Write will copy bytes from b into the pipe.
// Will block if buffer is full.
// Will return 0 if pipe is closed.
func (bp *BytePipe) Write(b []byte) int {
	if atomic.LoadUint32(&bp.s) == 0 {
		return 0
	}
	wr := atomic.LoadUint32(&bp.written)
	l := uint32(len(b))

	// fmt.Printf("WRITE: wants to write %d bytes starting at wi:%d.\n", l, wi)
	if wr == bp.cap {
		// fmt.Printf("WRITE: buffer full, wait for read.\n")
		// wait
		bp.wrCond.L.Lock()
		for bp.written == bp.cap {
			bp.wrCond.Wait()
		}
		wr = bp.written
		bp.wrCond.L.Unlock()
		if atomic.LoadUint32(&bp.s) == 0 {
			return 0
		}
	}

	// Now have SOME space
	if wr+l > bp.cap {
		l = bp.cap - wr
		// fmt.Printf("WRITE: wr: %d, l: %d\n", wr, l)
		// fmt.Printf("WRITE: not enough space to complete write, cut down to %d.\n", l)
	}
	// Now we know how much we CAN write
	var idx uint32
	// Check if what we have fits in one write

	// Start writing from start if we are at end.
	if bp.wi == bp.cap {
		bp.wi = 0
	}
	if bp.wi+l > bp.cap {
		// fmt.Printf("WRITE: write hit cap, rolling over to 0 idx.\n")
		// Can't fit everything at the end.
		idx = bp.cap - bp.wi
		copy(bp.wbuf[bp.wi:], b[:idx])
		// fmt.Printf("WRITE: buffer state: %v\n", bp.wbuf)
		// fmt.Printf("WRITE: writing %d bytes start from %d.\n", idx, wi)
		l -= idx
		bp.wi = 0
	}

	// write what what have
	copy(bp.wbuf[bp.wi:], b[idx:idx+l])
	// Store new write index
	// Update total written bytes
	bp.wi += l
	total := l + idx

	bp.wrCond.L.Lock()
	bp.written += total
	bp.wrCond.Broadcast()
	bp.wrCond.L.Unlock()

	// If we wrote everything, exit here
	if total == uint32(len(b)) {
		// fmt.Printf("WRITE: wrote all %d bytes of request: %d.\n", total, int(l))
		return int(total)
	}

	// fmt.Printf("WRITE: recursing to write %d to %d.\n", total, len(b))
	return int(total) + bp.Write(b[total:]) // Now write the rest
}

// Read will copy bytes from pipe into provided buffer.
// Will block if pipe is empty.
// Will read out as many bytes are availble (max of provided buffer size)
// Returns 0 if pipe is closed.
func (bp *BytePipe) Read(b []byte) int {
	if atomic.LoadUint32(&bp.s) == 0 {
		return 0
	}
	wr := atomic.LoadUint32(&bp.written)
	l := uint32(len(b))

	// If written == 0 there is nothing in pipe
	if wr == 0 {
		// fmt.Printf("READ: nothing to read, waiting.\n")
		// We dont have anything to read, wait for now
		bp.wrCond.L.Lock()
		for bp.written == 0 {
			bp.wrCond.Wait()
		}
		wr = bp.written
		bp.wrCond.L.Unlock()
		if atomic.LoadUint32(&bp.s) == 0 {
			return 0
		}
		wr = atomic.LoadUint32(&bp.written)
	}

	// We for sure have at least 1 byte
	// Set L to how much we can read
	if wr < l {
		// fmt.Printf("READ: only %d bytes available (wanted %d)\n", wr, l)
		l = wr
	}

	if bp.ri == bp.cap {
		bp.ri = 0
	}

	var total uint32 // how many bytes did we read?
	// Check if what we have fits in one read
	if bp.ri+l > bp.cap {
		total = bp.cap - bp.ri
		copy(b, bp.rbuf[bp.ri:bp.ri+total])
		// fmt.Printf("READ:  buffer state: %v\n", bp.rbuf)
		l -= total
		bp.ri = 0
		// fmt.Printf("READ: read %d bytes from end of buffer. left to read l: %d\n", idx, l)
	}

	// fmt.Printf("READ: Reading %d bytes.\n", l)
	// now read how much is left to read
	copy(b[total:], bp.rbuf[bp.ri:bp.ri+l])
	// fmt.Printf("READ:  buffer state: %v\n", bp.rbuf)
	bp.ri += l
	total += l
	// fmt.Printf("READ: total now: %d\n", total)
	// Update total written bytes
	bp.wrCond.L.Lock()
	bp.written -= total
	bp.wrCond.Broadcast()
	bp.wrCond.L.Unlock()
	// fmt.Printf("READ: bp.Written now at: %d\n", atomic.LoadUint32(&bp.written))
	// fmt.Printf("READ: Written decremented by %d. returning %d bytes read.\n", total, total)
	return int(total) // Now return how much we read
}

// Close will shutdown the pipe and signal any waiting goroutines.
func (bp *BytePipe) Close() {
	atomic.StoreUint32(&bp.s, 0)
	atomic.StoreUint32(&bp.written, 0)
}
