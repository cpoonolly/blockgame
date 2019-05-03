package main

import (
	"fmt"
	"syscall/js"

	"github.com/cpoonolly/blockgame/core"
	"github.com/cpoonolly/blockgame/wasm/webgl"
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

	/* Canvas Resize */

	onCanvasResize := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		game.OnViewPortChange()

		return nil
	})

	/* Key Press Tracking */

	isKeyDownMap := make(map[string]bool)
	wasKeyPressedMap := make(map[string]bool)

	onKeyDown := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = true

		fmt.Printf("Key Down %s\n", args[0].Get("code").String())

		return nil
	})
	onKeyUp := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = false
		wasKeyPressedMap[args[0].Get("code").String()] = true

		fmt.Printf("Key Up %s\n", args[0].Get("code").String())

		return nil
	})

	/* Main Game Loop */

	var lastRenderTime float32
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := float32(args[0].Float())
		dt := now - lastRenderTime
		lastRenderTime = now

		inputMap := map[core.GameInput]bool{
			core.GameInputPlayerMoveForward: false,
			core.GameInputPlayerMoveBack:    false,
			core.GameInputPlayerMoveLeft:    false,
			core.GameInputPlayerMoveRight:   false,
			core.GameInputCameraZoomIn:      false,
			core.GameInputCameraZoomOut:     false,
			core.GameInputCameraRotateLeft:  false,
			core.GameInputCameraRotateRight: false,
			core.GameInputEditModeToggle:    false,
			core.GameInputEditModeMoveUp:    false,
			core.GameInputEditModeMoveDown:  false,
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

		if wasKeyPressedMap["KeyQ"] {
			inputMap[core.GameInputEditModeToggle] = true
		}
		if isKeyDownMap["KeyA"] {
			inputMap[core.GameInputEditModeMoveUp] = true
		}
		if isKeyDownMap["KeyS"] {
			inputMap[core.GameInputEditModeMoveDown] = true
		}
		if isKeyDownMap["KeyW"] {
			inputMap[core.GameInputEditModeCreateWorldBlock] = true
		}
		if isKeyDownMap["KeyE"] {
			inputMap[core.GameInputEditModeCreateEnemy] = true
		}
		if isKeyDownMap["KeyD"] {
			inputMap[core.GameInputEditModeDelete] = true
		}

		game.Update(dt, inputMap)
		game.Render()

		js.Global().Call("requestAnimationFrame", renderFrame)
		clearMap(wasKeyPressedMap)

		if inputMap[core.GameInputEditModeToggle] {
			gl.DocumentEl.Call("getElementById", "container_main").Get("classList").Call("toggle", "edit-mode-enabled")
			game.OnViewPortChange()
		}

		if len(game.Log) > 0 || inputMap[core.GameInputEditModeToggle] {
			gl.DocumentEl.Call("getElementById", "game_log").Set("innerHTML", game.Log)
		}
		return nil
	})

	/* Editor Actions */

	exportGame := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		gl.DocumentEl.Call("getElementById", "import-export-val").Set("value", game.ExportAsJSON())

		return nil
	})

	importGame := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		gameData := gl.DocumentEl.Call("getElementById", "import-export-val").Get("value")

		if err := game.ImportFromJSON(gameData.String()); err != nil {
			panic(err)
		}

		return nil
	})

	defer renderFrame.Release()
	defer onKeyDown.Release()
	defer onKeyUp.Release()
	defer onCanvasResize.Release()
	defer exportGame.Release()
	defer importGame.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)
	js.Global().Call("addEventListener", "keydown", onKeyDown)
	js.Global().Call("addEventListener", "keyup", onKeyUp)
	js.Global().Call("addEventListener", "resize", onCanvasResize)
	js.Global().Set("exportGame", exportGame)
	js.Global().Set("importGame", importGame)

	done := make(chan struct{}, 0)
	<-done
}
