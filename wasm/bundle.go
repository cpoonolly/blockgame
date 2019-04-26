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
	enemyBlockIDs := make([]uint32, 0)

	renderEditBlockAttr := func(blockID uint32, label, attrName string, value float32) string {
		return fmt.Sprintf(`
			<div class='edit-block-attr'>
				<label class='edit-block-attr-label'>%[2]s:</label>
				<input id='edit-block-%[3]s-%[1]d' class='edit-block-attr-val' type='number' step='0.1' value='%.1[4]f'/>
			</div>
			`,
			blockID,
			label,
			attrName,
			value,
		)
	}

	renderEditBlockPanel := func(blockID uint32, blockType string, position, dimensions, color [3]float32) string {
		return fmt.Sprintf(`
				<div id='edit-block-%[1]d' class='edit-block'>
					<h5 class='edit-block-title'>%[2]s: %[1]d</h5><br/>
					%[3]s %[4]s %[5]s %[6]s %[7]s %[8]s %[9]s %[10]s %[11]s
					<button class='edit-block-update-btn' onclick='updateBlock(%[1]d, "%[2]s")'>Update</button>
					<button class='edit-block-delete-btn' onclick='deleteBlock(%[1]d, "%[2]s")'>Delete</button>
				</div>
			`,
			blockID,
			blockType,
			renderEditBlockAttr(blockID, "x", "posx", position[0]),
			renderEditBlockAttr(blockID, "y", "posy", position[1]),
			renderEditBlockAttr(blockID, "z", "posz", position[2]),
			renderEditBlockAttr(blockID, "width", "dimx", dimensions[0]),
			renderEditBlockAttr(blockID, "height", "dimy", dimensions[1]),
			renderEditBlockAttr(blockID, "length", "dimz", dimensions[2]),
			renderEditBlockAttr(blockID, "r", "colr", color[0]),
			renderEditBlockAttr(blockID, "g", "colg", color[1]),
			renderEditBlockAttr(blockID, "b", "colb", color[2]),
		)
	}

	renderEditorPanel := func() {
		var htmlBuilder strings.Builder

		htmlBuilder.WriteString(`
			<button class='new-block-btn' onclick='createNewBlock("World Block")'>New World Block</button>
			<button class='new-block-btn' onclick='createNewBlock("Enemy")'>Enemy</button>
			<div class='edit-blocks'>
		`)

		for _, worldBlockID := range worldBlockIDs {
			position := game.GetWorldBlockPosition(worldBlockID)
			dimensions := game.GetWorldBlockDimensions(worldBlockID)
			color := game.GetWorldBlockColor(worldBlockID)

			htmlBuilder.WriteString(renderEditBlockPanel(worldBlockID, "World Block", position, dimensions, color))
		}

		for _, enemyBlockID := range enemyBlockIDs {
			position := game.GetEnemyPosition(enemyBlockID)
			dimensions := game.GetEnemyDimensions(enemyBlockID)
			color := game.GetEnemyColor(enemyBlockID)

			htmlBuilder.WriteString(renderEditBlockPanel(enemyBlockID, "Enemy", position, dimensions, color))
		}

		htmlBuilder.WriteString("</div>")

		gl.DocumentEl.Call("getElementById", "container_editor_panel").Set("innerHTML", htmlBuilder.String())
	}

	createNewBlock := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		blockType := args[0].String()
		position := [3]float32{-1, -1, -1}
		dimensions := [3]float32{1, 1, 1}

		var newBlockID uint32
		if blockType == "World Block" {
			color := [3]float32{.7, .7, .7}
			newBlockID = game.CreateWorldBlock(position, dimensions, color)
			worldBlockIDs = append(worldBlockIDs, newBlockID)
		} else if blockType == "Enemy" {
			color := [3]float32{1, .3, .3}
			newBlockID = game.CreateEnemy(position, dimensions, color)
			enemyBlockIDs = append(enemyBlockIDs, newBlockID)
		}

		renderEditorPanel()

		return nil
	})

	updateBlock := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var position, dimensions, color [3]float32

		blockID := uint32(args[0].Int())
		blockType := args[1].String()

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

		color[0] = getAttribValue("colr")
		color[1] = getAttribValue("colg")
		color[2] = getAttribValue("colb")

		if blockType == "World Block" {
			game.UpdateWorldBlock(blockID, position, dimensions, color)
		} else if blockType == "Enemy" {
			game.UpdateEnemy(blockID, position, dimensions, color)
		}

		renderEditorPanel()

		return nil
	})

	deleteBlock := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		blockID := uint32(args[0].Int())
		blockType := args[1].String()

		if blockType == "World Block" {
			game.DeleteWorldBlock(blockID)

			// remove block from ids
			for index, worldBlockID := range worldBlockIDs {
				if worldBlockID == blockID {
					copy(worldBlockIDs[index:], worldBlockIDs[index+1:])
					worldBlockIDs = worldBlockIDs[:len(worldBlockIDs)-1]
					break
				}
			}
		} else if blockType == "Enemy" {
			game.DeleteEnemy(blockID)

			// remove block from ids
			for index, enemyBlockID := range enemyBlockIDs {
				if enemyBlockID == blockID {
					copy(enemyBlockIDs[index:], enemyBlockIDs[index+1:])
					enemyBlockIDs = enemyBlockIDs[:len(enemyBlockIDs)-1]
					break
				}
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
