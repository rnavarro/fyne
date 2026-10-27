//go:build !wasm && !test_web_driver

package glfw

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/build"

	"github.com/go-gl/glfw/v3.4/glfw"
)

// platform values returned by forcePlatform to override GLFW's auto-detection.
const (
	platformAuto    = ""
	platformX11     = "x11"
	platformWayland = "wayland"
)

// glfwState gates the calls that are only valid between glfw.Init and
// glfw.Terminate. runGL sets running before initGLFW, so a fyne.Do queued from
// another goroutine in that window (app.watchTheme does at startup) called
// PostEmptyEvent too early and panicked with NotInitialized.
var glfwState struct {
	sync.RWMutex
	ready bool
}

func (d *gLDriver) initGLFW() {
	switch forcePlatform() {
	case platformX11:
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	case platformWayland:
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformWayland))
	}

	err := glfw.Init()
	if err != nil {
		fyne.LogError("failed to initialise GLFW", err)
		return
	}
	glfwState.Lock()
	glfwState.ready = true
	glfwState.Unlock()

	initCursors()
	if glfw.GetPlatform() == glfw.PlatformWayland {
		build.IsWayland = true
	}
}

// pollEvents blocks for up to 100ms waiting for events, so an idle app does
// not busy-poll at 60fps. It does not block while fyne.Do work is queued:
// runGL's select picks the event tick over a ready funcQueue about half the
// time, and the PostEmptyEvent wake for that work may already have been
// consumed, so waiting here delayed queued work by 100-700ms. A function
// queued after the check increments pendingFuncs before posting its wake, so
// the wake is never lost.
//
// It does not block while a visible canvas is dirty either: drawSingleFrame
// runs after pollEvents in the same tick, so parking delayed a repaint by up
// to 100ms (an isolated indicator update took 101-118ms to reach the screen,
// 1-17ms without the wait).
func (d *gLDriver) pollEvents() {
	if pendingFuncs.Load() > 0 || d.visibleCanvasDirty() {
		glfw.PollEvents()
		return
	}
	glfw.WaitEventsTimeout(0.1)
}

// wakeEventLoop unblocks WaitEventsTimeout so the main loop promptly drains the
// funcQueue (fyne.Do work) or observes d.done. Safe to call from any goroutine:
// glfw.PostEmptyEvent is documented thread-safe.
// Before glfw.Init and after Terminate it does nothing; work queued before
// Init is drained by the first loop iteration, which polls instead of waiting
// while pendingFuncs is non-zero.
func wakeEventLoop() {
	glfwState.RLock()
	defer glfwState.RUnlock()
	if glfwState.ready {
		glfw.PostEmptyEvent()
	}
}

func (d *gLDriver) Terminate() {
	glfwState.Lock()
	defer glfwState.Unlock()
	glfwState.ready = false
	glfw.Terminate()
}
