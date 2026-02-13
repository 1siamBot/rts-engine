package network

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"
)

// LockstepManager synchronizes game ticks across network
type LockstepManager struct {
	mu           sync.Mutex
	localPlayer  int
	pendingCmds  map[uint64][]GameCommand // tick -> commands
	confirmedTick uint64
	inputDelay   int // ticks of input delay (typically 2-3)
	conn         *net.UDPConn
	remoteAddr   *net.UDPAddr
	isHost       bool
	connected    bool
}

func NewLockstepManager(localPlayer int, isHost bool) *LockstepManager {
	return &LockstepManager{
		localPlayer: localPlayer,
		pendingCmds: make(map[uint64][]GameCommand),
		inputDelay:  2,
		isHost:      isHost,
	}
}

// Host starts listening for connections
func (lm *LockstepManager) Host(port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	lm.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	lm.connected = true
	go lm.receiveLoop()
	return nil
}

// Join connects to a host
func (lm *LockstepManager) Join(host string, port int) error {
	remote, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	local, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}
	lm.conn, err = net.ListenUDP("udp", local)
	if err != nil {
		return err
	}
	lm.remoteAddr = remote
	lm.connected = true
	go lm.receiveLoop()
	return nil
}

// QueueCommand adds a local command to be sent on the scheduled tick
func (lm *LockstepManager) QueueCommand(currentTick uint64, cmd GameCommand) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	scheduledTick := currentTick + uint64(lm.inputDelay)
	cmd.Tick = scheduledTick
	lm.pendingCmds[scheduledTick] = append(lm.pendingCmds[scheduledTick], cmd)

	// Send to remote
	if lm.conn != nil && lm.remoteAddr != nil {
		var buf bytes.Buffer
		_ = cmd.Encode(&buf)
		_, _ = lm.conn.WriteToUDP(buf.Bytes(), lm.remoteAddr)
	}
}

// GetCommands returns all commands for a given tick
func (lm *LockstepManager) GetCommands(tick uint64) []GameCommand {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	cmds := lm.pendingCmds[tick]
	delete(lm.pendingCmds, tick)
	return cmds
}

// IsConnected returns true if network is active
func (lm *LockstepManager) IsConnected() bool {
	return lm.connected
}

func (lm *LockstepManager) receiveLoop() {
	buf := make([]byte, 4096)
	for {
		if lm.conn == nil {
			return
		}
		_ = lm.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, addr, err := lm.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if lm.isHost && lm.remoteAddr == nil {
			lm.remoteAddr = addr
		}

		var cmd GameCommand
		r := bytes.NewReader(buf[:n])
		if err := cmd.Decode(r); err != nil {
			continue
		}

		lm.mu.Lock()
		lm.pendingCmds[cmd.Tick] = append(lm.pendingCmds[cmd.Tick], cmd)
		lm.mu.Unlock()
	}
}

// Close shuts down the network connection
func (lm *LockstepManager) Close() {
	lm.connected = false
	if lm.conn != nil {
		lm.conn.Close()
	}
}
