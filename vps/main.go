package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Game struct {
	Board         [3][3]int
	Winner        int
	CurrentPlayer int
	mu            sync.Mutex
}

func NewGame() *Game {
	return &Game{
		Board:         [3][3]int{},
		Winner:        0,
		CurrentPlayer: 1,
	}
}

var game = NewGame()

func (g *Game) CheckGameState() {

	for i := 0; i < 3; i++ {
		if g.Board[i][0] == g.Board[i][1] && g.Board[i][1] == g.Board[i][2] && g.Board[i][0] != 0 ||
			g.Board[0][i] == g.Board[1][i] && g.Board[1][i] == g.Board[2][i] && g.Board[0][i] != 0 {
			g.Winner = g.CurrentPlayer
			return
		}
	}

	if g.Board[0][0] == g.Board[1][1] && g.Board[1][1] == g.Board[2][2] && g.Board[0][0] != 0 ||
		g.Board[0][2] == g.Board[1][1] && g.Board[1][1] == g.Board[2][0] && g.Board[0][2] != 0 {
		g.Winner = g.CurrentPlayer
		return
	}

	isDraw := true
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.Board[i][j] == 0 {
				isDraw = false
				break
			}
		}
	}

	if isDraw {
		g.Winner = 3
	}
}

func (g *Game) MakeMove(player, row, col int) bool {

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Winner != 0 || g.Board[row][col] != 0 || player != g.CurrentPlayer {
		return false
	}

	g.Board[row][col] = player
	g.CheckGameState()
	g.CurrentPlayer = 3 - player
	return true
}

func (g *Game) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Println("Invalid request method")
		return
	}

	var move struct {
		Player, Row, Col int
	}

	if err := json.NewDecoder(r.Body).Decode(&move); err != nil {
		log.Println("Failed to decode request body:", err)
		return
	}

	if ok := g.MakeMove(move.Player, move.Row, move.Col); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	game.mu.Lock()
	defer game.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.Board)
}

func main() {
	http.Handle("/game", game)
	http.HandleFunc("/state", stateHandler)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
