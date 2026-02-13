package network

import (
	"encoding/json"
	"fmt"
	"net"
)

// LobbyState represents the state of a game lobby
type LobbyState struct {
	HostName   string       `json:"host_name"`
	MapName    string       `json:"map_name"`
	MaxPlayers int          `json:"max_players"`
	Players    []LobbySlot  `json:"players"`
	Started    bool         `json:"started"`
	Port       int          `json:"port"`
}

// LobbySlot represents a player slot in the lobby
type LobbySlot struct {
	PlayerID int    `json:"player_id"`
	Name     string `json:"name"`
	Faction  string `json:"faction"`
	Team     int    `json:"team"`
	Ready    bool   `json:"ready"`
	IsAI     bool   `json:"is_ai"`
}

// Lobby manages pre-game setup
type Lobby struct {
	State    LobbyState
	listener net.Listener
	IsHost   bool
	ChatLog  []string
}

// NewLobby creates a new lobby as host
func NewLobby(hostName, mapName string, maxPlayers, port int) *Lobby {
	return &Lobby{
		State: LobbyState{
			HostName:   hostName,
			MapName:    mapName,
			MaxPlayers: maxPlayers,
			Port:       port,
			Players: []LobbySlot{
				{PlayerID: 0, Name: hostName, Faction: "Allied", Team: 0},
			},
		},
		IsHost: true,
	}
}

// AddPlayer adds a player to the lobby
func (l *Lobby) AddPlayer(name, faction string, isAI bool) int {
	id := len(l.State.Players)
	if id >= l.State.MaxPlayers {
		return -1
	}
	l.State.Players = append(l.State.Players, LobbySlot{
		PlayerID: id,
		Name:     name,
		Faction:  faction,
		Team:     id,
		IsAI:     isAI,
	})
	return id
}

// SetReady marks a player as ready
func (l *Lobby) SetReady(playerID int, ready bool) {
	for i := range l.State.Players {
		if l.State.Players[i].PlayerID == playerID {
			l.State.Players[i].Ready = ready
		}
	}
}

// AllReady checks if all players are ready
func (l *Lobby) AllReady() bool {
	for _, p := range l.State.Players {
		if !p.Ready && !p.IsAI {
			return false
		}
	}
	return len(l.State.Players) >= 2
}

// Chat adds a chat message
func (l *Lobby) Chat(playerName, msg string) {
	l.ChatLog = append(l.ChatLog, fmt.Sprintf("[%s] %s", playerName, msg))
}

// Marshal returns JSON of the lobby state
func (l *Lobby) Marshal() ([]byte, error) {
	return json.Marshal(l.State)
}
