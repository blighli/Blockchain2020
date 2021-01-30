package node

import (
	"github.com/pipapa/pbft/cmd"
	"github.com/pipapa/pbft/message"
)

type Sequence struct {
	lastSequence message.Sequence
	sequence     message.Sequence
	waterL       message.Sequence
	waterH       message.Sequence
}

func NewSequence(cfg *cmd.SharedConfig) *Sequence {
	return &Sequence{
		lastSequence: -1,
		sequence:     -1,
		waterL:       message.Sequence(cfg.WaterL),
		waterH:       message.Sequence(cfg.WaterH),
	}
}

func (s *Sequence) Get() message.Sequence {
	s.sequence = s.sequence + 1
	return s.sequence
}

func (s *Sequence) CheckBound(seq message.Sequence) bool {
	if seq < s.lastSequence {
		return false
	}
	if seq < s.waterL || seq > s.waterH {
		return false
	}
	return true
}

func (s *Sequence) SetLastSequence(sequence message.Sequence) {
	s.lastSequence = sequence
}

func (s *Sequence) GetLastSequence() message.Sequence {
	return s.lastSequence
}
