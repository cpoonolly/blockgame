package main

import (
	"fmt"
	"github.com/cpoonolly/blockgame/wasm/webgl"
	"github.com/go-gl/mathgl/mgl32"
	"syscall/js"
)

var (
	gl     *webgl.Context
	shader *webgl.ShaderProgram
	mesh   *webgl.Mesh
	err    error
)

func main() {
	fmt.Println("Let's give this a try...")

	gl, err = webgl.New("canvas_main")
	if err != nil {
		fmt.Println(err)
		return
	}

	body := gl.DocumentEl.Get("body")
	canvasWidth := body.Get("clientWidth").Int()
	gl.CanvasEl.Set("width", canvasWidth)
	canvasHeight := body.Get("clientHeight").Int()
	gl.CanvasEl.Set("height", canvasHeight)

	vertShaderCode := `
		attribute vec3 position;

		uniform mat4 pMatrix;
		uniform mat4 vMatrix;
		uniform mat4 mMatrix;
		uniform vec4 color;

		varying highp vec3 vColor;

		void main(void) {
			gl_Position = pMatrix * vMatrix * mMatrix * vec4(position, 1.);

			vColor = color.rgb;
		}	
	`
	fragShaderCode := `
		varying highp vec3 vColor;

		void main(void) {
			gl_FragColor = vec4(vColor, 1.);
		}
	`
	shader, err = gl.NewShaderProgram(vertShaderCode, fragShaderCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	verticies := []float32{
		-1, -1, -1, 1, -1, -1, 1, 1, -1, -1, 1, -1,
		-1, -1, 1, 1, -1, 1, 1, 1, 1, -1, 1, 1,
		-1, -1, -1, -1, 1, -1, -1, 1, 1, -1, -1, 1,
		1, -1, -1, 1, 1, -1, 1, 1, 1, 1, -1, 1,
		-1, -1, -1, -1, -1, 1, 1, -1, 1, 1, -1, -1,
		-1, 1, -1, -1, 1, 1, 1, 1, 1, 1, 1, -1,
	}
	normals := []float32{}
	indicies := []uint16{
		0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7,
		8, 9, 10, 8, 10, 11, 12, 13, 14, 12, 14, 15,
		16, 17, 18, 16, 18, 19, 20, 21, 22, 20, 22, 23,
	}

	mesh = gl.NewMesh(verticies, normals, indicies)

	projMatrix := mgl32.Perspective(mgl32.DegToRad(45.0), float32(canvasWidth)/float32(canvasHeight), 0, 1000.0)
	viewMatrix := mgl32.Ident4()
	modelMatrix := mgl32.Ident4()

	uniformsMat4f := map[string]js.TypedArray{
		"pMatrix": js.TypedArrayOf(projMatrix[:]),
		"vMatrix": js.TypedArrayOf(viewMatrix[:]),
		"mMatrix": js.TypedArrayOf(modelMatrix[:]),
	}

	colorVec := []float32{1.0, 1.0, 1.0, 1.0}
	uniformVec4f := map[string]js.TypedArray{
		"color": js.TypedArrayOf(colorVec[:]),
	}

	var cameraX, cameraY, cameraZ float32
	var renderFrame, onKeyDown, onKeyUp js.Func

	var isLeftPressed, isRightPressed, isUpPressed, isDownPressed, isCtrlPressed bool

	onKeyDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		fmt.Printf("key down - code: %s\n", event.Get("code"))
		keyCode := event.Get("code").String()

		if keyCode == "ArrowLeft" {
			isLeftPressed = true
		}
		if keyCode == "ArrowRight" {
			isRightPressed = true
		}
		if keyCode == "ArrowUp" {
			isUpPressed = true
		}
		if keyCode == "ArrowDown" {
			isDownPressed = true
		}
		if keyCode == "ControlLeft" {
			isCtrlPressed = true
		}

		return nil
	})

	onKeyUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		fmt.Printf("key up - code: %s\n", event.Get("code"))
		keyCode := event.Get("code").String()

		if keyCode == "ArrowLeft" {
			isLeftPressed = false
		}
		if keyCode == "ArrowRight" {
			isRightPressed = false
		}
		if keyCode == "ArrowUp" {
			isUpPressed = false
		}
		if keyCode == "ArrowDown" {
			isDownPressed = false
		}
		if keyCode == "ControlLeft" {
			isCtrlPressed = false
		}

		return nil
	})

	var lastRenderTime float32
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := float32(args[0].Float())
		dt := now - lastRenderTime
		lastRenderTime = now

		if isLeftPressed {
			cameraX -= 0.01
		}
		if isRightPressed {
			cameraX += 0.01
		}

		if isCtrlPressed {
			if isUpPressed {
				cameraY += 0.01
			}
			if isDownPressed {
				cameraY -= 0.01
			}
		} else {
			if isUpPressed {
				cameraZ -= 0.01
			}
			if isDownPressed {
				cameraZ += 0.01
			}
		}

		gl.DocumentEl.Call("getElementById", "fps_counter").Set("innerHTML", fmt.Sprintf("FPS: %f", 1000.0/dt))

		viewMatrix = mgl32.LookAtV(mgl32.Vec3{3.0 + cameraX, 3.0 + cameraY, 3.0 + cameraZ}, mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, 1.0, 0.0})
		modelMatrix = mgl32.Ident4()

		gl.ClearScreen()
		gl.Render(mesh, shader, uniformsMat4f, uniformVec4f)

		js.Global().Call("requestAnimationFrame", renderFrame)

		return nil
	})

	defer renderFrame.Release()
	defer onKeyDown.Release()
	defer onKeyUp.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)
	js.Global().Call("addEventListener", "keydown", onKeyDown)
	js.Global().Call("addEventListener", "keyup", onKeyUp)

	fmt.Println("Did it work???")

	done := make(chan struct{}, 0)
	<-done
}
