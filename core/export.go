package core

import (
	"encoding/json"
)

type blockData struct {
	Position [3]float32
	Color    [3]float32
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
	for _, worldBlock := range game.worldBlocks {
		var worldBlockData blockData

		worldBlockData.Position = worldBlock.pos
		worldBlockData.Color = worldBlock.color.Vec3()

		data.World = append(data.World, worldBlockData)
	}

	data.Enemies = make([]blockData, 0, len(game.enemies))
	for _, enemy := range game.enemies {
		var enemyData blockData

		enemyData.Position = enemy.pos
		enemyData.Color = enemy.color.Vec3()

		data.World = append(data.World, enemyData)
	}

	json, _ := json.Marshal(&data)

	return string(json)
}
