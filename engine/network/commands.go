package network

import (
	"encoding/binary"
	"io"
)

// CmdType identifies a network command
type CmdType uint8

const (
	CmdMoveUnit CmdType = iota
	CmdAttackUnit
	CmdStopUnit
	CmdBuildUnit
	CmdPlaceBuilding
	CmdSellBuilding
	CmdSetRally
	CmdChat
)

// GameCommand is a deterministic command that modifies game state
type GameCommand struct {
	Tick     uint64
	PlayerID int
	Type     CmdType
	EntityID uint64
	TargetX  int32
	TargetY  int32
	Param    string // unit type name, chat text, etc.
}

// Encode writes a command to binary
func (c *GameCommand) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, c.Tick); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, int32(c.PlayerID)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, c.Type); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, c.EntityID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, c.TargetX); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, c.TargetY); err != nil {
		return err
	}
	paramBytes := []byte(c.Param)
	if err := binary.Write(w, binary.LittleEndian, uint16(len(paramBytes))); err != nil {
		return err
	}
	_, err := w.Write(paramBytes)
	return err
}

// Decode reads a command from binary
func (c *GameCommand) Decode(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &c.Tick); err != nil {
		return err
	}
	var pid int32
	if err := binary.Read(r, binary.LittleEndian, &pid); err != nil {
		return err
	}
	c.PlayerID = int(pid)
	if err := binary.Read(r, binary.LittleEndian, &c.Type); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.EntityID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.TargetX); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.TargetY); err != nil {
		return err
	}
	var plen uint16
	if err := binary.Read(r, binary.LittleEndian, &plen); err != nil {
		return err
	}
	if plen > 0 {
		buf := make([]byte, plen)
		if _, err := io.ReadFull(r, buf); err != nil {
			return err
		}
		c.Param = string(buf)
	}
	return nil
}
