package glhf

import "github.com/go-gl/gl/v3.3-core/gl"

// Init initializes OpenGL by loading function pointers from the active OpenGL context.
// This function must be manually run inside the main thread (using "github.com/faiface/mainthread"
// package).
//
// It must be called under the presence of an active OpenGL context, e.g., always after calling
// window.MakeContextCurrent(). Also, always call this function when switching contexts.
func Init() {
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	gl.Enable(gl.BLEND)
	gl.Enable(gl.SCISSOR_TEST)
	gl.BlendEquation(gl.FUNC_ADD)
}

// Clear clears the current framebuffer or window with the given color.
func Clear(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

// Bounds sets the drawing bounds in pixels. Drawing outside bounds is always discarted.
//
// Calling this function is equivalent to setting viewport and scissor in OpenGL.
func Bounds(x, y, w, h int) {
	gl.Viewport(int32(x), int32(y), int32(w), int32(h))
	gl.Scissor(int32(x), int32(y), int32(w), int32(h))
}

// BlendFactor represents a source or destination blend factor.
type BlendFactor int

// Here's the list of all blend factors.
const (
	One              = BlendFactor(gl.ONE)
	Zero             = BlendFactor(gl.ZERO)
	SrcAlpha         = BlendFactor(gl.SRC_ALPHA)
	DstAlpha         = BlendFactor(gl.DST_ALPHA)
	OneMinusSrcAlpha = BlendFactor(gl.ONE_MINUS_SRC_ALPHA)
	OneMinusDstAlpha = BlendFactor(gl.ONE_MINUS_DST_ALPHA)
)

// BlendFunc sets the source and destination blend factor.
func BlendFunc(src, dst BlendFactor) {
	gl.BlendFunc(uint32(src), uint32(dst))
}
