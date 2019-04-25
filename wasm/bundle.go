package main

import (
	"fmt"
	"github.com/cpoonolly/blockgame/core"
	"github.com/cpoonolly/blockgame/wasm/webgl"
	"strconv"
	"strings"
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

	onCanvasResize := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		game.OnViewPortChange()

		return nil
	})

	isKeyDownMap := make(map[string]bool)
	wasKeyPressedMap := make(map[string]bool)

	onKeyDown := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = true

		// fmt.Printf("KeyDown: %s\n", args[0].Get("code").String())

		return nil
	})
	onKeyUp := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = false
		wasKeyPressedMap[args[0].Get("code").String()] = true

		// fmt.Printf("KeyUp: %s\n", args[0].Get("code").String())
		// fmt.Printf("KeyPressed: %s\n", args[0].Get("code").String())

		return nil
	})

	var lastRenderTime float32
	var renderFrame js.Func
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

		if inputMap[core.GameInputEditModeToggle] {
			gl.DocumentEl.Call("getElementById", "container_main").Get("classList").Call("toggle", "edit-mode-enabled")
			game.OnViewPortChange()
		}

		if len(game.Log) > 0 || inputMap[core.GameInputEditModeToggle] {
			gl.DocumentEl.Call("getElementById", "game_log").Set("innerHTML", game.Log)
		}
		return nil
	})

	worldBlockIDs := make([]uint32, 0)

	renderEditorPanel := func() {
		var htmlBuilder strings.Builder

		htmlBuilder.WriteString("<button class='' onclick='createNewBlock()'>New Block</button><br/><br/>")
		htmlBuilder.WriteString("<div class='edit-blocks'>")

		for _, worldBlockID := range worldBlockIDs {
			position := game.GetWorldBlockPosition(worldBlockID)
			dimensions := game.GetWorldBlockDimensions(worldBlockID)

			htmlBuilder.WriteString(fmt.Sprintf("<div id='edit-block-%d' class='edit-block'>", worldBlockID))
			htmlBuilder.WriteString(fmt.Sprintf("<strong>Block: %d</strong><br/>", worldBlockID))
			htmlBuilder.WriteString(fmt.Sprintf("<span>x:<input id='edit-block-posx-%d' type='number' value='%.2f'/></span><br/>", worldBlockID, position[0]))
			htmlBuilder.WriteString(fmt.Sprintf("<span>y:<input id='edit-block-posy-%d' type='number' value='%.2f'/></span><br/>", worldBlockID, position[1]))
			htmlBuilder.WriteString(fmt.Sprintf("<span>z:<input id='edit-block-posz-%d' type='number' value='%.2f'/></span><br/>", worldBlockID, position[2]))
			htmlBuilder.WriteString(fmt.Sprintf("<span>width:<input id='edit-block-dimx-%d' type='number' value='%.2f'/></span><br/>", worldBlockID, dimensions[0]))
			htmlBuilder.WriteString(fmt.Sprintf("<span>height:<input id='edit-block-dimy-%d' type='number' value='%.2f'/></span><br/>", worldBlockID, dimensions[1]))
			htmlBuilder.WriteString(fmt.Sprintf("<span>length:<input id='edit-block-dimz-%d' type='number' value='%.2f'/></span><br/>", worldBlockID, dimensions[2]))
			htmlBuilder.WriteString(fmt.Sprintf("<button onclick='updateBlock(%d)'>Update</button>", worldBlockID))
			htmlBuilder.WriteString("</div>")
		}

		htmlBuilder.WriteString("</div>")

		gl.DocumentEl.Call("getElementById", "container_editor_panel").Set("innerHTML", htmlBuilder.String())
	}

	createNewBlock := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		position := [3]float32{-1, -1, -1}
		dimensions := [3]float32{1, 1, 1}
		color := [3]float32{.7, .7, .7}

		newBlockID := game.EditorCreateWorldBlock(position, dimensions, color)
		worldBlockIDs = append(worldBlockIDs, newBlockID)

		renderEditorPanel()

		return nil
	})

	updateBlock := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var position, dimensions, color [3]float32

		blockID := uint32(args[0].Int())

		getAttribValue := func(attrName string) float32 {
			attrID := fmt.Sprintf("edit-block-%s-%d", attrName, blockID)
			attrEl := gl.DocumentEl.Call("getElementById", attrID)
			attrVal, err := strconv.ParseFloat(attrEl.Get("value").String(), 32)
			if err != nil {
				fmt.Println(fmt.Errorf("invalid attribute %s: %v", attrName, err))
				attrVal = 0
			}

			return float32(attrVal)
		}

		position[0] = getAttribValue("posx")
		position[1] = getAttribValue("posy")
		position[2] = getAttribValue("posz")

		dimensions[0] = getAttribValue("dimx")
		dimensions[1] = getAttribValue("dimy")
		dimensions[2] = getAttribValue("dimz")

		color = [3]float32{.7, .7, .7}

		game.EditorUpdateWorldBlock(blockID, position, dimensions, color)
		renderEditorPanel()

		return nil
	})

	defer renderFrame.Release()
	defer onKeyDown.Release()
	defer onKeyUp.Release()
	defer onCanvasResize.Release()
	defer createNewBlock.Release()
	defer updateBlock.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)
	js.Global().Call("addEventListener", "keydown", onKeyDown)
	js.Global().Call("addEventListener", "keyup", onKeyUp)
	js.Global().Call("addEventListener", "resize", onCanvasResize)
	js.Global().Set("createNewBlock", createNewBlock)
	js.Global().Set("updateBlock", updateBlock)

	renderEditorPanel()

	fmt.Println("Did it work???")

	done := make(chan struct{}, 0)
	<-done
}
