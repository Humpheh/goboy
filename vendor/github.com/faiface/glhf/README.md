# glhf [![GoDoc](https://godoc.org/github.com/faiface/glhf?status.svg)](http://godoc.org/github.com/faiface/glhf) [![Report card](https://goreportcard.com/badge/github.com/faiface/glhf)](https://goreportcard.com/report/github.com/faiface/glhf)

open**GL** **H**ave **F**un - A Go package that makes life with OpenGL enjoyable.

```
go get github.com/faiface/glhf
```

## Main features

- Garbage collected OpenGL objects
- Dynamically sized vertex slices (vertex arrays are boring)
- Textures, Shaders, Frames (reasonably managed framebuffers)
- Always possible to use standard OpenGL with `glhf`

## Motivation

OpenGL is verbose, it's usage patterns are repetitive and it's manual memory management doesn't fit
Go's design. When making a game development library, it's usually desirable to create some
higher-level abstractions around OpenGL. This library is a take on that.

## Contribute!

The library is young and many features are still missing. If you find a bug, have a proposal or a
feature request, _do an issue_!. If you know how to implement something that's missing, _do a pull
request_.

## Code

The following are parts of the demo program, which can be found in the [examples](https://github.com/faiface/glhf/tree/master/examples/demo).

```go
// ... GLFW window creation and stuff ...

// vertex shader source
var vertexShader = `
#version 330 core

in vec2 position;
in vec2 texture;

out vec2 Texture;

void main() {
	gl_Position = vec4(position, 0.0, 1.0);
	Texture = texture;
}
`

// fragment shader source
var fragmentShader = `
#version 330 core

in vec2 Texture;

out vec4 color;

uniform sampler2D tex;

void main() {
	color = texture(tex, Texture);
}
`

var (
        // Here we define a vertex format of our vertex slice. It's actually a basic slice
        // literal.
        //
        // The vertex format consists of names and types of the attributes. The name is the
        // name that the attribute is referenced by inside a shader.
        vertexFormat = glhf.AttrFormat{
                {Name: "position", Type: glhf.Vec2},
                {Name: "texture", Type: glhf.Vec2},
        }

        // Here we declare some variables for later use.
        shader  *glhf.Shader
        texture *glhf.Texture
        slice   *glhf.VertexSlice
)

// Here we load an image from a file. The loadImage function is not within the library, it
// just loads and returns a image.NRGBA.
gopherImage, err := loadImage("celebrate.png")
if err != nil {
        panic(err)
}

// Every OpenGL call needs to be done inside the main thread.
mainthread.Call(func() {
        var err error

        // Here we create a shader. The second argument is the format of the uniform
        // attributes. Since our shader has no uniform attributes, the format is empty.
        shader, err = glhf.NewShader(vertexFormat, glhf.AttrFormat{}, vertexShader, fragmentShader)

        // If the shader compilation did not go successfully, an error with a full
        // description is returned.
        if err != nil {
                panic(err)
        }

        // We create a texture from the loaded image.
        texture = glhf.NewTexture(
                gopherImage.Bounds().Dx(),
                gopherImage.Bounds().Dy(),
                true,
                gopherImage.Pix,
        )

        // And finally, we make a vertex slice, which is basically a dynamically sized
        // vertex array. The length of the slice is 6 and the capacity is the same.
        //
        // The slice inherits the vertex format of the supplied shader. Also, it should
        // only be used with that shader.
        slice = glhf.MakeVertexSlice(shader, 6, 6)

        // Before we use a slice, we need to Begin it. The same holds for all objects in
        // GLHF.
        slice.Begin()

        // We assign data to the vertex slice. The values are in the order as in the vertex
        // format of the slice (shader). Each two floats correspond to an attribute of type
        // glhf.Vec2.
        slice.SetVertexData([]float32{
                -1, -1, 0, 1,
                +1, -1, 1, 1,
                +1, +1, 1, 0,

                -1, -1, 0, 1,
                +1, +1, 1, 0,
                -1, +1, 0, 0,
        })

        // When we're done with the slice, we End it.
        slice.End()
})

shouldQuit := false
for !shouldQuit {
        mainthread.Call(func() {
                // ... GLFW stuff ...

                // Clear the window.
                glhf.Clear(1, 1, 1, 1)

                // Here we Begin/End all necessary objects and finally draw the vertex
                // slice.
                shader.Begin()
                texture.Begin()
                slice.Begin()
                slice.Draw()
                slice.End()
                texture.End()
                shader.End()

                // ... GLFW stuff ...
        })
}
```

## FAQ

### Which version of OpenGL does GLHF use?

It uses OpenGL 3.3 and uses
[`github.com/go-gl/gl/v3.3-core/gl`](https://github.com/go-gl/gl/tree/master/v3.3-core/gl).

### Why do I have to use `github.com/faiface/mainthread` package with GLHF?

First of all, OpenGL has to be done from one thread and many operating systems require, that the one
thread will be the main thread of your application.

But why that specific package? GLHF uses the `mainthread` package to do the garbage collection of
OpenGL objects, which is super convenient. So in order for it to work correctly, you have to
initialize the `mainthread` package through `mainthread.Run`. However, once you call this function
there is no way to run functions on the main thread, except for through the `mainthread` package.

### Why is the important XY feature not included?

I probably didn't need it yet. If you want that features, create an issue or implement it and do a
pull request.

### Does GLHF create windows for me?

No. You have to use another library for windowing, e.g.
[github.com/go-gl/glfw/v3.2/glfw](https://github.com/go-gl/glfw/tree/master/v3.2/glfw).

### Why no tests?

If you find a way to automatically test OpenGL, I may add tests.