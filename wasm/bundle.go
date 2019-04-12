package main

import (
	"fmt"
	"github.com/cpoonolly/blockgame/wasm/webgl"
)

var (
	gl     *webgl.Context
	shader *webgl.ShaderProgram
	mesh   *webgl.Mesh
	err    error
)

func main() {
	// Creating a channel so that wasm will keep running
	c := make(chan struct{}, 0)

	fmt.Println("Let's give this a try...")

	gl, err = webgl.New("canvas_main")
	if err != nil {
		fmt.Print(err)
		return
	}

	body := gl.DocumentEl.Get("body")
	canvasWidth := body.Get("clientWidth").Int()
	gl.CanvasEl.Set("width", canvasWidth)
	canvasHeight := body.Get("clientHeight").Int()
	gl.CanvasEl.Set("height", canvasHeight)

	vertShaderCode := `
		attribute vec3 position;
				
		void main(void) {
			gl_Position = vec4(position, 1.0);
		}
	`
	fragShaderCode := `
		void main(void) {
			gl_FragColor = vec4(0.0, 0.0, 1.0, 1.0);
		}
	`
	shader, err = gl.NewShaderProgram(vertShaderCode, fragShaderCode)
	if err != nil {
		fmt.Print(err)
		return
	}

	indicies := []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
	}
	elements := []uint32{
		2, 1, 0,
	}
	mesh = gl.NewMesh(indicies, elements)

	gl.ClearScreen()
	gl.Render(mesh, shader, nil, nil)

	fmt.Println("Did it work???")

	<-c
}
