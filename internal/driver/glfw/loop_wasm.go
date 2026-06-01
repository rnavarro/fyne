//go:build wasm || test_web_driver

package glfw

import (
	"fyne.io/fyne/v2"

	"github.com/fyne-io/gl-js"
	"github.com/fyne-io/glfw-js"
)

func (d *gLDriver) initGLFW() {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		fyne.LogError("failed to initialise GLFW", err)
		return
	}
}

func (d *gLDriver) pollEvents() {
	glfw.PollEvents() // This call blocks while window is being resized, which prevents freeDirtyTextures from being called
}

// wakeEventLoop is a no-op on wasm: the browser event loop is not blocked on a
// WaitEventsTimeout call, so there is nothing to wake. Present so shared code
// (driver.Quit, runOnMainWithWait) compiles on all platforms.
func wakeEventLoop() {}

func (d *gLDriver) Terminate() {
	glfw.Terminate()
}
