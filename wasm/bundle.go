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

func clearMap(m map[string]bool) {
	for k := range m {
		delete(m, k)
	}
}

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
	wasKeyPressedMap := make(map[string]bool)

	onKeyDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = true
		fmt.Printf("KeyDown: %s\n", args[0].Get("code").String())
		return nil
	})
	onKeyUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = false
		wasKeyPressedMap[args[0].Get("code").String()] = true

		fmt.Printf("KeyUp: %s\n", args[0].Get("code").String())
		fmt.Printf("KeyPressed: %s\n", args[0].Get("code").String())

		return nil
	})

	var lastRenderTime float32
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := float32(args[0].Float())
		dt := now - lastRenderTime
		lastRenderTime = now

		inputMap := map[core.GameInput]bool{
			core.GameInputPlayerMoveForward:   false,
			core.GameInputPlayerMoveBack:      false,
			core.GameInputPlayerMoveLeft:      false,
			core.GameInputPlayerMoveRight:     false,
			core.GameInputCameraZoomIn:        false,
			core.GameInputCameraZoomOut:       false,
			core.GameInputCameraRotateLeft:    false,
			core.GameInputCameraRotateRight:   false,
			core.GameInputEditModeToggle:      false,
			core.GameInputEditModeMoveUp:      false,
			core.GameInputEditModeMoveDown:    false,
			core.GameInputEditModeBlockCreate: false,
			core.GameInputEditModeBlockDelete: false,
		}

		if isKeyDownMap["ControlLeft"] {
			if isKeyDownMap["ArrowUp"] {
				inputMap[core.GameInputCameraZoomIn] = true
			}
			if isKeyDownMap["ArrowDown"] {
				inputMap[core.GameInputCameraZoomOut] = true
			}
			if isKeyDownMap["ArrowLeft"] {
				inputMap[core.GameInputCameraRotateLeft] = true
			}
			if isKeyDownMap["ArrowRight"] {
				inputMap[core.GameInputCameraRotateRight] = true
			}
		} else {
			if isKeyDownMap["ArrowUp"] {
				inputMap[core.GameInputPlayerMoveForward] = true
			}
			if isKeyDownMap["ArrowDown"] {
				inputMap[core.GameInputPlayerMoveBack] = true
			}
			if isKeyDownMap["ArrowLeft"] {
				inputMap[core.GameInputPlayerMoveLeft] = true
			}
			if isKeyDownMap["ArrowRight"] {
				inputMap[core.GameInputPlayerMoveRight] = true
			}
		}

		if wasKeyPressedMap["KeyE"] {
			inputMap[core.GameInputEditModeToggle] = true
		}
		if wasKeyPressedMap["KeyW"] {
			inputMap[core.GameInputEditModeBlockCreate] = true
		}
		if wasKeyPressedMap["KeyD"] {
			inputMap[core.GameInputEditModeBlockDelete] = true
		}
		if isKeyDownMap["KeyA"] {
			inputMap[core.GameInputEditModeMoveUp] = true
		}
		if isKeyDownMap["KeyS"] {
			inputMap[core.GameInputEditModeMoveDown] = true
		}

		game.Update(dt, inputMap)
		game.Render()

		js.Global().Call("requestAnimationFrame", renderFrame)
		clearMap(wasKeyPressedMap)

		// gl.DocumentEl.Call("getElementById", "game_log").Set("innerHTML", game.Log)
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
