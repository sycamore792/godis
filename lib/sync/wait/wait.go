package wait

import (
	"sync"
	"time"
)

type WaitingGroup struct {
	wg sync.WaitGroup
}

// Add adds delta, which may be negative, to the WaitingGroup counter.
func (w *WaitingGroup) Add(delta int) {
	w.wg.Add(delta)
}

// Done decrements the WaitingGroup counter by one
func (w *WaitingGroup) Done() {
	w.wg.Done()
}

// WaitingGroup blocks until the WaitingGroup counter is zero.
func (w *WaitingGroup) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout blocks until the WaitingGroup counter is zero or timeout
// returns true if timeout
func (w *WaitingGroup) WaitWithTimeout(timeout time.Duration) bool {
	// 这个缓冲通道能不能去掉呢
	c := make(chan bool, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
