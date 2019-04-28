package core

import (
	"encoding/json"
	"github.com/go-gl/mathgl/mgl32"
)

type blockData struct {
	Position   [3]float32
	Dimensions [3]float32
	Color      [3]float32
}

type gameData struct {
	Player  blockData
	World   []blockData
	Enemies []blockData
}

// ExportAsJSON exports the game into json data
func (game *Game) ExportAsJSON() string {
	var data gameData

	data.Player.Position = game.player.pos
	data.Player.Color = game.player.color.Vec3()

	data.World = make([]blockData, 0, len(game.worldBlocks))
	for worldBlockID := range game.worldBlocks {
		var worldBlockData blockData

		worldBlockData.Position = game.GetWorldBlockPosition(worldBlockID)
		worldBlockData.Dimensions = game.GetWorldBlockDimensions(worldBlockID)
		worldBlockData.Color = game.GetWorldBlockColor(worldBlockID)

		data.World = append(data.World, worldBlockData)
	}

	data.Enemies = make([]blockData, 0, len(game.enemies))
	for enemyID := range game.enemies {
		var enemyData blockData

		enemyData.Position = game.GetEnemyPosition(enemyID)
		enemyData.Dimensions = game.GetEnemyDimensions(enemyID)
		enemyData.Color = game.GetEnemyColor(enemyID)

		data.World = append(data.World, enemyData)
	}

	json, _ := json.Marshal(&data)

	return string(json)
}

// ImportFromJSON imports the game from json data. returns worldBlockIDs enemyIds or an error
func (game *Game) ImportFromJSON(jsonData string) ([]uint32, []uint32, error) {
	var data gameData

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, nil, err
	}

	game.player.pos = data.Player.Position
	game.player.color = mgl32.Vec3(data.Player.Color).Vec4(1.0)

	worldBlockIDs := make([]uint32, 0, len(data.World))
	game.worldBlocks = make(map[uint32]*worldBlock)
	for _, worldBlockData := range data.World {
		worldBlockID := game.CreateWorldBlock(worldBlockData.Position, worldBlockData.Dimensions, worldBlockData.Color)
		worldBlockIDs = append(worldBlockIDs, worldBlockID)
	}

	enemyIDs := make([]uint32, 0, len(data.Enemies))
	game.enemies = make(map[uint32]*enemy)
	for _, enemyData := range data.Enemies {
		enemyID := game.CreateEnemy(enemyData.Position, enemyData.Dimensions, enemyData.Color)
		enemyIDs = append(enemyIDs, enemyID)
	}

	return worldBlockIDs, enemyIDs, nil
}
