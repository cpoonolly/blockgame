package main

import (
	"fmt"
	"github.com/cpoonolly/blockgame/core"
	"github.com/cpoonolly/blockgame/wasm/webgl"
	"syscall/js"
)

var (
	gl   *webgl.Context
	game *core.Game
	err  error
)

func main() {
	fmt.Println("Let's give this a try...")

	gl, err = webgl.New("canvas_main")
	if err != nil {
		fmt.Println(err)
		return
	}

	game, err := core.NewGame(gl)
	if err != nil {
		fmt.Println(err)
		return
	}

	var cameraX, cameraY, cameraZ float32
	var renderFrame, onKeyDown, onKeyUp js.Func

	isKeyDownMap := make(map[string]bool)
	onKeyDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = true
		return nil
	})
	onKeyUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = false
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

		game.Update(cameraX, cameraY, cameraZ)
		game.Render()

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
