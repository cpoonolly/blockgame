package main

import (
	"fmt"
	"strconv"
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

		// fmt.Printf("Key Down %s\n", args[0].Get("code").String())

		return nil
	})
	onKeyUp := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = false
		wasKeyPressedMap[args[0].Get("code").String()] = true

		// fmt.Printf("Key Up %s\n", args[0].Get("code").String())

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
			if isKeyDownMap["Space"] {
				inputMap[core.GameInputPlayerJump] = true
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

	movePlayerTo := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		moveToX, errX := strconv.ParseFloat(gl.DocumentEl.Call("getElementById", "move-to-x").Get("value").String(), 32)
		if errX != nil {
			panic(errX)
		}

		moveToY, errY := strconv.ParseFloat(gl.DocumentEl.Call("getElementById", "move-to-y").Get("value").String(), 32)
		if errY != nil {
			panic(errY)
		}

		moveToZ, errZ := strconv.ParseFloat(gl.DocumentEl.Call("getElementById", "move-to-z").Get("value").String(), 32)
		if errZ != nil {
			panic(errZ)
		}

		game.MovePlayerToPos([3]float32{float32(moveToX), float32(moveToY), float32(moveToZ)})

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
	js.Global().Set("movePlayerTo", movePlayerTo)

	defaultMap := "{\"player\":{\"position\":[17.0, 5.0, 17.0],\"dimensions\":[1,1,1]},\"world\":[{\"position\":[0,0,0],\"dimensions\":[30,0.5,30]},{\"position\":[0,0,0],\"dimensions\":[1,5,31]},{\"position\":[1,0,0],\"dimensions\":[29,5,1]},{\"position\":[30,0,0],\"dimensions\":[1,5,31]},{\"position\":[1,0,30],\"dimensions\":[29,5,1]},{\"position\":[15,0,15],\"dimensions\":[5,3,5]},{\"position\":[19.333858,4.663249,10.518786],\"dimensions\":[2.3844757,0.9663763,3.4307919]},{\"position\":[19.81797,8.043748,16.667824],\"dimensions\":[3.157837,0.5,3.0037613]},{\"position\":[13.755774,10.925252,14.243277],\"dimensions\":[3.1519737,0.5,3.1608505]},{\"position\":[19.752327,13.4904995,14.212866],\"dimensions\":[3.1099472,0.5,3.5069046]},{\"position\":[13.141777,19.019375,17.493816],\"dimensions\":[3.5883484,0.5,3.2169342]},{\"position\":[17.71711,17.070627,17.4342],\"dimensions\":[1.3740082,0.5,1.8798332]},{\"position\":[9.9834385,19.7851,14.854664],\"dimensions\":[1.8013802,0.5,1.9732056]},{\"position\":[10.680285,20.484118,10.934053],\"dimensions\":[1.4222565,0.5,1.374588]},{\"position\":[10.817757,21.283535,5.738801],\"dimensions\":[1.3983421,0.5,1.9267006]},{\"position\":[11.764454,22.365986,0.5622523],\"dimensions\":[1.5518188,0.5,1.9198413]},{\"position\":[13.898621,25.346954,-4.2526617],\"dimensions\":[2.3314896,0.5,3.1563582]},{\"position\":[13.890339,27.095861,-19.547745],\"dimensions\":[0.47509003,0.5,13.060982]},{\"position\":[9.817467,29.427332,-28.7391],\"dimensions\":[5.102867,0.5,6.3680305]},{\"position\":[10.350336,31.758835,-33.195694],\"dimensions\":[2.472643,0.5,1.8093109]},{\"position\":[9.632517,33.24095,-38.226765],\"dimensions\":[2.1125278,0.5,2.8368073]},{\"position\":[7.1313553,0.5,6.5342093],\"dimensions\":[3.1799088,6.0781703,2.6655798]},{\"position\":[6.7412844,0.5,22.67184],\"dimensions\":[1.9776316,6.810938,2.4174194]},{\"position\":[23.095049,0.5,20.4995],\"dimensions\":[1.9512405,7.0273113,2.0497665]},{\"position\":[23.919891,0.5,7.150091],\"dimensions\":[2.399582,7.560281,2.5389977]}],\"enemies\":[{\"position\":[25,2,5],\"dimensions\":[1,1,1]}]}"
	if err := game.ImportFromJSON(defaultMap); err != nil {
		panic(err)
	}

	done := make(chan struct{}, 0)
	<-done
}
