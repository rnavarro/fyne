//go:build !wasm && !test_web_driver

package glfw

import (
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
func (d *gLDriver) pollEvents() {
	if pendingFuncs.Load() > 0 {
		glfw.PollEvents()
		return
	}
	glfw.WaitEventsTimeout(0.1)
}

// wakeEventLoop unblocks WaitEventsTimeout so the main loop promptly drains the
// funcQueue (fyne.Do work) or observes d.done. Safe to call from any goroutine:
// glfw.PostEmptyEvent is documented thread-safe.
func wakeEventLoop() {
	glfw.PostEmptyEvent()
}

func (d *gLDriver) Terminate() {
	glfw.Terminate()
}
