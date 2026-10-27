//go:build !no_glfw && !mobile && !wasm && !test_web_driver

package glfw

import (
	"testing"
	"time"
)

// Queued main-thread work must not wait out the event loop's idle timeout.
// A caller posting one function every 20 ms (an audio level meter) saw
// 100-600 ms delays when the loop parked in WaitEventsTimeout while work was
// still queued.
func TestRunOnMain_QueuedWorkDoesNotWaitForIdleTimeout(t *testing.T) {
	const calls = 40
	slow := 0
	var worst time.Duration
	for i := 0; i < calls; i++ {
		queued := time.Now()
		ran := make(chan time.Duration, 1)
		runOnMainWithWait(func() { ran <- time.Since(queued) }, false)
		lat := <-ran
		if lat > 50*time.Millisecond {
			slow++
		}
		if lat > worst {
			worst = lat
		}
		time.Sleep(20 * time.Millisecond)
	}
	if slow > 2 {
		t.Errorf("%d of %d queued calls waited more than 50ms (worst %v)", slow, calls, worst)
	}
}
