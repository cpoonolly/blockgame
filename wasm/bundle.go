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

		return nil
	})
	onKeyUp := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isKeyDownMap[args[0].Get("code").String()] = false
		wasKeyPressedMap[args[0].Get("code").String()] = true

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

	/* Editor Actions */

	worldBlockIDs := make([]uint32, 0)

	renderEditorPanel := func() {
		var htmlBuilder strings.Builder

		htmlBuilder.WriteString(`
			<button class='new-block-btn' onclick='createNewBlock()'>New Block</button>
			<div class='edit-blocks'>
		`)

		for _, worldBlockID := range worldBlockIDs {
			position := game.GetWorldBlockPosition(worldBlockID)
			dimensions := game.GetWorldBlockDimensions(worldBlockID)

			htmlBuilder.WriteString(fmt.Sprintf(`
					<div id='edit-block-%[1]d' class='edit-block'>
						<h5 class='edit-block-title'>Block: %[1]d</h5><br/>
						<div class='edit-block-attr'>
							<label class='edit-block-attr-label'>x:</label>
							<input id='edit-block-posx-%[1]d' class='edit-block-attr-val' type='number' value='%.2[2]f'/>
						</div>
						<div class='edit-block-attr'>
							<label class='edit-block-attr-label'>y:</label>
							<input id='edit-block-posy-%[1]d' class='edit-block-attr-val' type='number' value='%.2[3]f'/>
						</div>
						<div class='edit-block-attr'>
							<label class='edit-block-attr-label'>z:</label>
							<input id='edit-block-posz-%[1]d' class='edit-block-attr-val' type='number' value='%.2[4]f'/>
						</div>
						<div class='edit-block-attr'>
							<label class='edit-block-attr-label'>width:</label>
							<input id='edit-block-posx-%[1]d' class='edit-block-attr-val' type='number' value='%.2[5]f'/>
						</div>
						<div class='edit-block-attr'>
							<label class='edit-block-attr-label'>height:</label>
							<input id='edit-block-posy-%[1]d' class='edit-block-attr-val' type='number' value='%.2[6]f'/>
						</div>
						<div class='edit-block-attr'>
							<label class='edit-block-attr-label'>length:</label>
							<input id='edit-block-posz-%[1]d' class='edit-block-attr-val' type='number' value='%.2[7]f'/>
						</div>
						<button class='edit-block-update-btn' onclick='updateBlock(%[1]d)'>Update</button>
						<button class='edit-block-delete-btn' onclick='deleteBlock(%[1]d)'>Delete</button>
					</div>
				`,
				worldBlockID,
				position[0],
				position[1],
				position[2],
				dimensions[0],
				dimensions[1],
				dimensions[2],
			))
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

	deleteBlock := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		blockID := uint32(args[0].Int())

		game.EditorDeleteWorldBlock(blockID)

		// remove block from ids
		for index, worldBlockID := range worldBlockIDs {
			if worldBlockID == blockID {
				copy(worldBlockIDs[index:], worldBlockIDs[index+1:])
				worldBlockIDs = worldBlockIDs[:len(worldBlockIDs)-1]
				break
			}
		}

		renderEditorPanel()

		return nil
	})

	defer renderFrame.Release()
	defer onKeyDown.Release()
	defer onKeyUp.Release()
	defer onCanvasResize.Release()
	defer createNewBlock.Release()
	defer updateBlock.Release()
	defer deleteBlock.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)
	js.Global().Call("addEventListener", "keydown", onKeyDown)
	js.Global().Call("addEventListener", "keyup", onKeyUp)
	js.Global().Call("addEventListener", "resize", onCanvasResize)
	js.Global().Set("createNewBlock", createNewBlock)
	js.Global().Set("updateBlock", updateBlock)
	js.Global().Set("deleteBlock", deleteBlock)

	renderEditorPanel()

	fmt.Println("Did it work???")

	done := make(chan struct{}, 0)
	<-done
}
