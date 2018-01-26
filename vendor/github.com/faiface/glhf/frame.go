package glhf

import (
	"runtime"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
)

// Frame is a fixed resolution texture that you can draw on.
type Frame struct {
	fb, rf, df binder // framebuffer, read framebuffer, draw framebuffer
	tex        *Texture
}

// NewFrame creates a new fully transparent Frame with given dimensions in pixels.
func NewFrame(width, height int, smooth bool) *Frame {
	f := &Frame{
		fb: binder{
			restoreLoc: gl.FRAMEBUFFER_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindFramebuffer(gl.FRAMEBUFFER, obj)
			},
		},
		rf: binder{
			restoreLoc: gl.READ_FRAMEBUFFER_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindFramebuffer(gl.READ_FRAMEBUFFER, obj)
			},
		},
		df: binder{
			restoreLoc: gl.DRAW_FRAMEBUFFER_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, obj)
			},
		},
		tex: NewTexture(width, height, smooth, make([]uint8, width*height*4)),
	}

	gl.GenFramebuffers(1, &f.fb.obj)

	f.fb.bind()
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, f.tex.tex.obj, 0)
	f.fb.restore()

	runtime.SetFinalizer(f, (*Frame).delete)

	return f
}

func (f *Frame) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteFramebuffers(1, &f.fb.obj)
	})
}

// ID returns the OpenGL framebuffer ID of this Frame.
func (f *Frame) ID() uint32 {
	return f.fb.obj
}

// Begin binds the Frame. All draw operations will target this Frame until End is called.
func (f *Frame) Begin() {
	f.fb.bind()
}

// End unbinds the Frame. All draw operations will go to whatever was bound before this Frame.
func (f *Frame) End() {
	f.fb.restore()
}

// Blit copies rectangle (sx0, sy0, sx1, sy1) in this Frame onto rectangle (dx0, dy0, dx1, dy1) in
// dst Frame.
//
// If the dst Frame is nil, the destination will be the framebuffer 0, which is the screen.
//
// If the sizes of the rectangles don't match, the source will be stretched to fit the destination
// rectangle. The stretch will be either smooth or pixely according to the source Frame's
// smoothness.
func (f *Frame) Blit(dst *Frame, sx0, sy0, sx1, sy1, dx0, dy0, dx1, dy1 int) {
	f.rf.obj = f.fb.obj
	if dst != nil {
		f.df.obj = dst.fb.obj
	} else {
		f.df.obj = 0
	}
	f.rf.bind()
	f.df.bind()

	filter := gl.NEAREST
	if f.tex.smooth {
		filter = gl.LINEAR
	}

	gl.BlitFramebuffer(
		int32(sx0), int32(sy0), int32(sx1), int32(sy1),
		int32(dx0), int32(dy0), int32(dx1), int32(dy1),
		gl.COLOR_BUFFER_BIT, uint32(filter),
	)

	f.rf.restore()
	f.df.restore()
}

// Texture returns the Frame's underlying Texture that the Frame draws on.
func (f *Frame) Texture() *Texture {
	return f.tex
}
