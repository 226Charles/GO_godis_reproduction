package timewheel

import (
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	ch := make(chan time.Time)
	beginTime := time.Now()
	//print(ch)
	//fmt.Println(beginTime)
	//print("b")
	Delay(time.Second, "", func() {
		ch <- time.Now()
	})
	//print("a")
	execAt := <-ch
	delayDuration := execAt.Sub(beginTime)
	// usually 1.0~2.0 s
	if delayDuration < time.Second || delayDuration > 3*time.Second {
		t.Error("wrong execute time")
	}
}
