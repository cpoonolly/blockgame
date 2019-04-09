package main

import (
	"fmt"
	"github.com/cpoonolly/blockgame/wasm/gl"
)

func main() {
	fmt.Println("Let's give this a try...")

	webgl, err := gl.New("canvas_main")
	if err != nil {
		fmt.Print(err)
		return
	}

	body := webgl.DocumentEl.Get("body")
	canvasWidth := body.Get("clientWidth").Int()
	webgl.CanvasEl.Set("width", canvasWidth)
	canvasHeight := body.Get("clientHeight").Int()
	webgl.CanvasEl.Set("height", canvasHeight)

	webgl.Test()
	fmt.Println("Did it work???")
}
