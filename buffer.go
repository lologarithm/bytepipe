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

	if wr == bp.cap {
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

	// How much can we write?
	if wr+l > bp.cap {
		l = bp.cap - wr
	}

	// Start writing from start if we are at end.
	if bp.wi == bp.cap {
		bp.wi = 0
	}

	// Check if what we have fits in one write
	var idx uint32
	if bp.wi+l > bp.cap {
		// Can't fit everything at the end.
		idx = bp.cap - bp.wi
		copy(bp.wbuf[bp.wi:], b[:idx])
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
	bp.written += total // Update written total inside lock.
	bp.wrCond.Broadcast()
	bp.wrCond.L.Unlock()

	// If we wrote everything, exit here
	if total == uint32(len(b)) {
		return int(total)
	}
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

	// Set L to wr if we don't have enough data to fill b
	if wr < l {
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
		l -= total
		bp.ri = 0
	}

	// now read how much is left to read
	copy(b[total:], bp.rbuf[bp.ri:bp.ri+l])
	bp.ri += l
	total += l

	// Update total written bytes
	bp.wrCond.L.Lock()
	bp.written -= total
	bp.wrCond.Broadcast()
	bp.wrCond.L.Unlock()
	return int(total) // Now return how much we read
}

// Close will shutdown the pipe and signal any waiting goroutines.
func (bp *BytePipe) Close() {
	atomic.StoreUint32(&bp.s, 0)
	atomic.StoreUint32(&bp.written, 0)
}
