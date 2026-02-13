package network

import (
	"bufio"
	"os"
)

// Replay records and plays back game commands for replay
type Replay struct {
	Commands []GameCommand
	file     *os.File
	writer   *bufio.Writer
}

// NewReplayRecorder creates a replay file for recording
func NewReplayRecorder(path string) (*Replay, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &Replay{
		file:   f,
		writer: bufio.NewWriter(f),
	}, nil
}

// Record writes a command to the replay file
func (r *Replay) Record(cmd GameCommand) error {
	r.Commands = append(r.Commands, cmd)
	return cmd.Encode(r.writer)
}

// Close flushes and closes the replay file
func (r *Replay) Close() error {
	if r.writer != nil {
		r.writer.Flush()
	}
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// LoadReplay loads a replay file
func LoadReplay(path string) (*Replay, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	replay := &Replay{}
	reader := bufio.NewReader(f)
	for {
		var cmd GameCommand
		if err := cmd.Decode(reader); err != nil {
			break
		}
		replay.Commands = append(replay.Commands, cmd)
	}
	return replay, nil
}

// CommandsForTick returns all commands at a given tick during playback
func (r *Replay) CommandsForTick(tick uint64) []GameCommand {
	var result []GameCommand
	for _, c := range r.Commands {
		if c.Tick == tick {
			result = append(result, c)
		}
	}
	return result
}
