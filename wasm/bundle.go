package main

import (
	"fmt"
	"github.com/cpoonolly/blockgame/wasm/webgl"
)

var (
	gl  *webgl.Context
	err error
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
		attribute vec3 coordinates;
				
		void main(void) {
			gl_Position = vec4(coordinates, 1.0);
		}
	`

	fragShaderCode := `
		void main(void) {
			gl_FragColor = vec4(0.0, 0.0, 1.0, 1.0);
		}
	`

	gl.NewShaderProgram("my_shader", vertShaderCode, fragShaderCode)

	fmt.Println("Did it work???")

	<-c
}
