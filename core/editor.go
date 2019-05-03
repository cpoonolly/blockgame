package core

import "fmt"
import "github.com/go-gl/mathgl/mgl32"

const editorActionDebounce = 1000

type gameEditor struct {
	timeSinceLastAction float32
	startPos            mgl32.Vec3
	worldBlock          *worldBlock // world block currently being created in edit mode
	enemy               *enemy      // enemy block currently being created in edit mode
}

func (editor *gameEditor) update(game *Game, dt float32, inputs map[GameInput]bool) {
	if !game.IsEditModeEnabled {
		return
	}

	if editor.worldBlock != nil {
		editor.updateWorldBlock(game)
	}

	if editor.enemy != nil {
		editor.updateEnemy(game)
	}

	editor.timeSinceLastAction = editor.timeSinceLastAction + dt
	if editor.timeSinceLastAction < editorActionDebounce {
		return
	}

	if inputs[GameInputEditModeCreateWorldBlock] {
		// on first input of create world block we start creating on second we finish
		if editor.worldBlock == nil {
			editor.createWorldBlockStart(game)
		} else {
			editor.createWorldBlockEnd(game)
		}

		editor.timeSinceLastAction = 0
	}

	if inputs[GameInputEditModeCreateEnemy] {
		if editor.enemy == nil {
			editor.createEnemyStart(game)
		} else {
			editor.createEnemyEnd(game)
		}

		editor.timeSinceLastAction = 0
	}

	if inputs[GameInputEditModeDelete] {
		editor.deleteBlocks(game)

		editor.timeSinceLastAction = 0
	}
}

func (editor *gameEditor) render(game *Game, viewMatrix mgl32.Mat4) error {
	if editor.worldBlock != nil {
		if err := editor.worldBlock.render(game, viewMatrix); err != nil {
			return err
		}
	}

	if editor.enemy != nil {
		if err := editor.enemy.render(game, viewMatrix); err != nil {
			return err
		}
	}

	return nil
}

func (editor *gameEditor) updateWorldBlock(game *Game) {
	// we want the new right top front corner of the world block to be at the players left bottom back corner
	leftTopFront := mgl32.Vec3{
		f32Max(editor.startPos.X(), game.player.right()),
		f32Max(editor.startPos.Y(), game.player.bottom()),
		f32Max(editor.startPos.Z(), game.player.back()),
	}
	rightBottomBack := mgl32.Vec3{
		f32Min(editor.startPos.X(), game.player.right()),
		f32Min(editor.startPos.Y(), game.player.bottom()),
		f32Min(editor.startPos.Z(), game.player.back()),
	}
	widthHeightLength := leftTopFront.Add(rightBottomBack.Mul(-1.0))

	editor.worldBlock.scale = widthHeightLength.Mul(0.5)
	editor.worldBlock.pos = rightBottomBack.Add(editor.worldBlock.scale)
}

func (editor *gameEditor) updateEnemy(game *Game) {
	enemy := editor.enemy
	player := game.player

	// enemy should always just be directly behind the player
	enemy.start = player.pos.Add(mgl32.Vec3{0.0, 0.0, -2.0 * enemy.scale.Z()})
	enemy.pos = enemy.start
}

func (editor *gameEditor) createWorldBlockStart(game *Game) {
	fmt.Printf("create world block start (timeSinceLastAction: %.5f)\n", editor.timeSinceLastAction)

	editor.enemy = nil // only should be creating a world block or enemy at one given time

	// create the currently editing world block at the players current postions
	editor.worldBlock = new(worldBlock)
	editor.worldBlock.pos = game.player.pos
	editor.worldBlock.scale = mgl32.Vec3{0.0, 0.0, 0.0} // scale changes as we move the player
	editor.worldBlock.color = worldBlockColorDefault
	editor.startPos = game.player.pos
}

func (editor *gameEditor) createWorldBlockEnd(game *Game) {
	fmt.Printf("create world block end (timeSinceLastAction: %.5f)\n", editor.timeSinceLastAction)

	game.worldBlocks = append(game.worldBlocks, editor.worldBlock)
	editor.worldBlock = nil
}

func (editor *gameEditor) createEnemyStart(game *Game) {
	fmt.Printf("create enemy start (timeSinceLastAction: %.5f)\n", editor.timeSinceLastAction)

	editor.worldBlock = nil // only should be creating a world block or enemy at one given time

	editor.enemy = new(enemy)
	editor.enemy.scale = game.player.scale
	editor.enemy.color = enemyColorDefault
	editor.updateEnemy(game)

	editor.startPos = game.player.pos
}

func (editor *gameEditor) createEnemyEnd(game *Game) {
	fmt.Printf("create enemy end (timeSinceLastAction: %.5f)\n", editor.timeSinceLastAction)

	game.enemies = append(game.enemies, editor.enemy)
	editor.enemy = nil
}

func (editor *gameEditor) deleteBlocks(game *Game) {
	fmt.Printf("deleting blocks (timeSinceLastAction: %.5f)\n", editor.timeSinceLastAction)

	player := game.player
	worldBlocks := game.worldBlocks
	enemies := game.enemies

	worldBlocksNewLen := 0
	for i := 0; i < len(worldBlocks); i++ {
		if !checkForStaticOnStaticCollision(player, worldBlocks[i]) {
			worldBlocks[worldBlocksNewLen] = worldBlocks[i]
			worldBlocksNewLen++
		}
	}
	fmt.Printf("len(worldBlocks): %d\nnew len(worldBlocks): %d\n", len(worldBlocks), worldBlocksNewLen)
	game.worldBlocks = worldBlocks[:worldBlocksNewLen]

	enemiesNewLen := 0
	for i := 0; i < len(enemies); i++ {
		if !checkForStaticOnStaticCollision(player, enemies[i]) {
			enemies[enemiesNewLen] = enemies[i]
			enemiesNewLen++
		}
	}
	fmt.Printf("len(enemies): %d\nnew len(enemies): %d\n", len(enemies), enemiesNewLen)
	game.enemies = enemies[:enemiesNewLen]
}
