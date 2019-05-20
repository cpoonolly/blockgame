package core

import (
	"encoding/json"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

type blockData struct {
	Position   [3]float32 `json:"position"`
	Dimensions [3]float32 `json:"dimensions"`
}

type gameData struct {
	Player  blockData   `json:"player"`
	World   []blockData `json:"world"`
	Enemies []blockData `json:"enemies"`
}

func getBlockPosition(block collidable) mgl32.Vec3 {
	return mgl32.Vec3{block.right(), block.bottom(), block.back()}
}

func getBlockPosFromData(data blockData) mgl32.Vec3 {
	return mgl32.Vec3(data.Position).Add(getBlockScaleFromData(data))
}

func getBlockDimensions(block collidable) mgl32.Vec3 {
	return mgl32.Vec3{block.left() - block.right(), block.top() - block.bottom(), block.front() - block.back()}
}

func getBlockScaleFromData(data blockData) mgl32.Vec3 {
	return mgl32.Vec3(data.Dimensions).Mul(0.5)
}

// ExportAsJSON exports the game into json data
func (game *Game) ExportAsJSON() string {
	var data gameData

	data.Player.Position = getBlockDimensions(game.player)
	data.Player.Dimensions = getBlockDimensions(game.player)

	data.World = make([]blockData, 0, len(game.worldBlocks))
	for _, worldBlock := range game.worldBlocks {
		var worldBlockData blockData

		worldBlockData.Position = getBlockPosition(worldBlock)
		worldBlockData.Dimensions = getBlockDimensions(worldBlock)

		data.World = append(data.World, worldBlockData)
	}

	data.Enemies = make([]blockData, 0, len(game.enemies))
	for _, enemy := range game.enemies {
		var enemyData blockData

		enemyData.Position = getBlockPosition(enemy)
		enemyData.Dimensions = getBlockDimensions(enemy)

		data.Enemies = append(data.Enemies, enemyData)
	}

	json, _ := json.Marshal(&data)

	return string(json)
}

// ImportFromJSON imports the game from json data. returns worldBlockIDs enemyIds or an error
func (game *Game) ImportFromJSON(jsonData string) error {
	var data gameData

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}

	game.player.pos = getBlockPosFromData(data.Player)

	fmt.Printf("Imported Player - Pos: {x: %.2f, y: %.2f, z: %.2f}\n", game.player.pos.X(), game.player.pos.Y(), game.player.pos.Z())

	game.worldBlocks = make([]*worldBlock, 0, len(data.World))
	for _, worldBlockData := range data.World {
		worldBlock := new(worldBlock)

		worldBlock.pos = getBlockPosFromData(worldBlockData)
		worldBlock.scale = getBlockScaleFromData(worldBlockData)

		game.worldBlocks = append(game.worldBlocks, worldBlock)

		fmt.Printf("Imported WorldBlock - Pos: {x: %.2f, y: %.2f, z: %.2f} - Scale: {x: %.2f, y: %.2f, z: %.2f}\n", worldBlock.pos.X(), worldBlock.pos.Y(), worldBlock.pos.Z(), worldBlock.scale.X(), worldBlock.scale.Y(), worldBlock.scale.Z())
	}

	game.enemies = make([]*enemy, 0, len(data.Enemies))
	for _, enemyData := range data.Enemies {
		enemy := new(enemy)

		enemy.pos = getBlockPosFromData(enemyData)
		enemy.scale = getBlockScaleFromData(enemyData)
		enemy.start = enemy.pos

		game.enemies = append(game.enemies, enemy)

		fmt.Printf("Imported Enemy - Pos: {x: %.2f, y: %.2f, z: %.2f} - Scale: {x: %.2f, y: %.2f, z: %.2f}\n", enemy.pos.X(), enemy.pos.Y(), enemy.pos.Z(), enemy.scale.X(), enemy.scale.Y(), enemy.scale.Z())
	}

	return nil
}
