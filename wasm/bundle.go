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
		attribute vec3 normal;

		uniform mat4 pMatrix;
		uniform mat4 mvMatrix;
		uniform vec4 color;

		varying highp vec3 vLighting;
		varying highp vec3 vColor;

		void main(void) {
			gl_Position = pMatrix * mvMatrix * vec4(position, 1.);

			highp vec3 ambientLight = vec3(0.3, 0.3, 0.3);
      highp vec3 directionalLightColor = vec3(.5, .5, .5);
			highp vec3 directionalVector = normalize(vec3(0.85, 0.8, 0.75));
			highp float directional = max(dot(normal, directionalVector), 0.0);
			vLighting = ambientLight + (directionalLightColor * directional);

			vColor = color.rgb;
		}	
	`
	fragShaderCode := `
		varying highp vec3 vLighting;
		varying highp vec3 vColor;

		void main(void) {
			gl_FragColor = vec4(vColor * vLighting, 1.);
		}
	`
	shader, err = gl.NewShaderProgram(vertShaderCode, fragShaderCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	verticies := []float32{
		// Front face
		-1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
		-1.0, 1.0, 1.0,

		// Back face
		-1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0,
		1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,

		// Top face
		-1.0, 1.0, -1.0,
		-1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, -1.0,

		// Bottom face
		-1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		1.0, -1.0, 1.0,
		-1.0, -1.0, 1.0,

		// Right face
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,
		1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,

		// Left face
		-1.0, -1.0, -1.0,
		-1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,
	}
	normals := []float32{
		// Front
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,

		// Back
		0.0, 0.0, -1.0,
		0.0, 0.0, -1.0,
		0.0, 0.0, -1.0,
		0.0, 0.0, -1.0,

		// Top
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,

		// Bottom
		0.0, -1.0, 0.0,
		0.0, -1.0, 0.0,
		0.0, -1.0, 0.0,
		0.0, -1.0, 0.0,

		// Right
		1.0, 0.0, 0.0,
		1.0, 0.0, 0.0,
		1.0, 0.0, 0.0,
		1.0, 0.0, 0.0,

		// Left
		-1.0, 0.0, 0.0,
		-1.0, 0.0, 0.0,
		-1.0, 0.0, 0.0,
		-1.0, 0.0, 0.0,
	}
	indicies := []uint16{
		// front
		0, 1, 2,
		0, 2, 3,
		// back
		4, 5, 6,
		4, 6, 7,
		// top
		8, 9, 10,
		8, 10, 11,
		// bottom
		12, 13, 14,
		12, 14, 15,
		// right
		16, 17, 18,
		16, 18, 19,
		// left
		20, 21, 22,
		20, 22, 23,
	}

	mesh = gl.NewMesh(verticies, normals, indicies)

	projMatrix := mgl32.Perspective(mgl32.DegToRad(45.0), float32(canvasWidth)/float32(canvasHeight), 1, 50.0)
	viewMatrix := mgl32.Ident4()
	modelMatrix := mgl32.Ident4()
	modelViewMatrix := mgl32.Ident4()

	uniformsMat4f := map[string]js.TypedArray{
		"pMatrix":  js.TypedArrayOf(projMatrix[:]),
		"mvMatrix": js.TypedArrayOf(modelViewMatrix[:]),
	}

	colorVec := []float32{1.0, 1.0, 1.0, 1.0}
	uniformVec4f := map[string]js.TypedArray{
		"color": js.TypedArrayOf(colorVec[:]),
	}

	var cameraX, cameraY, cameraZ float32
	var renderFrame, onKeyDown, onKeyUp js.Func

	isKeyDownMap := make(map[string]bool)
	onKeyDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		fmt.Printf("key down - code: %s\n", event.Get("code"))
		keyCode := event.Get("code").String()
		isKeyDownMap[keyCode] = true

		return nil
	})
	onKeyUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		fmt.Printf("key up - code: %s\n", event.Get("code"))
		keyCode := event.Get("code").String()
		isKeyDownMap[keyCode] = false

		return nil
	})

	var lastRenderTime float32
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := float32(args[0].Float())
		dt := now - lastRenderTime
		lastRenderTime = now

		if isKeyDownMap["ArrowLeft"] {
			cameraX -= 0.1
		}
		if isKeyDownMap["ArrowRight"] {
			cameraX += 0.1
		}

		if isKeyDownMap["ControlLeft"] {
			if isKeyDownMap["ArrowUp"] {
				cameraY += 0.1
			}
			if isKeyDownMap["ArrowDown"] {
				cameraY -= 0.1
			}
		} else {
			if isKeyDownMap["ArrowUp"] {
				cameraZ -= 0.1
			}
			if isKeyDownMap["ArrowDown"] {
				cameraZ += 0.1
			}
		}

		gl.DocumentEl.Call("getElementById", "fps_counter").Set("innerHTML", fmt.Sprintf("FPS: %f", 1000.0/dt))

		viewMatrix = mgl32.LookAtV(mgl32.Vec3{0.0 + cameraX, 0.0 + cameraY, -6.0 + cameraZ}, mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, 1.0, 0.0})
		modelMatrix = mgl32.Ident4()
		modelViewMatrix = viewMatrix.Mul4(modelMatrix)

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
