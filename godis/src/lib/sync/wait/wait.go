package wait

/*sync.WaitGroup没有提供等待超时的功能
这个wait.Wait结构体提供了一个带超时的Wait方法，它能够等待一组线程的结束或超时
*/

import (
	"sync"
	"time"
)

type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

func (w *Wait) Done() {
	w.wg.Done()
}

func (w *Wait) Wait() {
	w.wg.Wait()
}

func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		w.Wait()
		c <- struct{}{}
	}()

	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}

}
