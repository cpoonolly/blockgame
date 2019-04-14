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

		var dx, dy, dz float32

		if isKeyDownMap["ControlLeft"] {
			if isKeyDownMap["ArrowUp"] {
				dy = 0.1
			}
			if isKeyDownMap["ArrowDown"] {
				dy = -0.1
			}
		} else {
			if isKeyDownMap["ArrowUp"] {
				dz = 0.1
			}
			if isKeyDownMap["ArrowDown"] {
				dz = -0.1
			}
			if isKeyDownMap["ArrowLeft"] {
				dx = 0.1
			}
			if isKeyDownMap["ArrowRight"] {
				dx = -0.1
			}
		}

		game.Update(dt, dx, dy, dz)
		game.Render()

		gl.DocumentEl.Call("getElementById", "game_log").Set("innerHTML", game.Log)
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
