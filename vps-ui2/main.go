package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"log"
	"net/http"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 300
	screenHeight = 300
	gridSize     = 100
)

var (
	gameState      Game
	currentPlayer  = 1
	serverEndpoint = "http://localhost:8080/game"
)

type Game struct {
	Board         [3][3]int
	Winner        int
	CurrentPlayer int
}

type GameScreen struct{}

func (g *GameScreen) Update() error {
	if gameState.Winner != 0 {
		return nil
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		row, col := y/gridSize, x/gridSize
		if row >= 0 && row < 3 && col >= 0 && col < 3 {
			makeMove(currentPlayer, row, col)
		}
	}
	fetchBoardState()
	return nil
}

func (g *GameScreen) Draw(screen *ebiten.Image) {
	// 绘制棋盘
	for i := 1; i <= 2; i++ {
		ebitenutil.DrawLine(screen, 0, float64(i*gridSize), screenWidth, float64(i*gridSize), color.White)
		ebitenutil.DrawLine(screen, float64(i*gridSize), 0, float64(i*gridSize), screenHeight, color.White)
	}

	// 绘制棋子
	for i, row := range gameState.Board {
		for j, cell := range row {
			if cell != 0 {
				cellStr := "X"
				if cell == 2 {
					cellStr = "O"
				}
				x := float64(j*gridSize) + gridSize/2
				y := float64(i*gridSize) + gridSize/2
				ebitenutil.DebugPrintAt(screen, cellStr, int(x), int(y))
			}
		}
	}

	if gameState.Winner != 0 {
		msg := "平局！"
		if gameState.Winner != 3 {
			msg = fmt.Sprintf("玩家%d胜利！", gameState.Winner)
		}
		ebitenutil.DebugPrintAt(screen, msg, 100, screenHeight/2)
	}
}

func (g *GameScreen) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	//println(outsideWidth, outsideHeight)
	return 300, 300
}

func makeMove(player, row, col int) {
	move := struct {
		Player, Row, Col int
	}{player, row, col}
	data, _ := json.Marshal(move)
	resp, err := http.Post(serverEndpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Failed to send move: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&gameState)
	}
}

func fetchBoardState() {
	resp, err := http.Get("http://localhost:8080/state")
	if err != nil {
		log.Printf("Failed to fetch board state: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var newBoard [3][3]int
	if err := json.NewDecoder(resp.Body).Decode(&newBoard); err != nil {
		log.Printf("Error decoding board state: %v\n", err)
	} else {
		gameState.Board = newBoard 
	}
}

func main() {
	game := &GameScreen{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("井字棋")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
