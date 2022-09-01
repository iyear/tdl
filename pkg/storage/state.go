package storage

import (
	"encoding/json"
	"errors"
	"github.com/gotd/td/telegram/updates"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
)

type State struct {
	kv *kv.KV
}

func NewState(kv *kv.KV) *State {
	return &State{kv: kv}
}

func (s *State) Get(key string, v interface{}) error {
	data, err := s.kv.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (s *State) Set(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return s.kv.Set(key, data)
}

func (s *State) GetState(userID int64) (updates.State, bool, error) {
	state := updates.State{}

	if err := s.Get(key.State(userID), &state); err != nil {
		if errors.Is(err, kv.ErrNotFound) {
			return state, false, nil
		}
		return state, false, err
	}

	return state, true, nil
}

func (s *State) SetState(userID int64, state updates.State) error {
	if err := s.Set(key.State(userID), state); err != nil {
		return err
	}

	return s.Set(key.StateChannel(userID), struct{}{})
}

func (s *State) SetPts(userID int64, pts int) error {
	state, k := updates.State{}, key.State(userID)

	if err := s.Get(k, &state); err != nil {
		return err
	}
	state.Pts = pts
	return s.Set(k, state)
}

func (s *State) SetQts(userID int64, qts int) error {
	state, k := updates.State{}, key.State(userID)

	if err := s.Get(k, &state); err != nil {
		return err
	}
	state.Qts = qts
	return s.Set(k, state)
}

func (s *State) SetDate(userID int64, date int) error {
	state, k := updates.State{}, key.State(userID)

	if err := s.Get(k, &state); err != nil {
		return err
	}
	state.Date = date
	return s.Set(k, state)
}

func (s *State) SetSeq(userID int64, seq int) error {
	state, k := updates.State{}, key.State(userID)

	if err := s.Get(k, &state); err != nil {
		return err
	}
	state.Seq = seq
	return s.Set(k, state)
}

func (s *State) SetDateSeq(userID int64, date, seq int) error {
	state, k := updates.State{}, key.State(userID)

	if err := s.Get(k, &state); err != nil {
		return err
	}
	state.Date = date
	state.Seq = seq
	return s.Set(k, state)
}

func (s *State) GetChannelPts(userID, channelID int64) (int, bool, error) {
	c := make(map[int64]int)

	if err := s.Get(key.StateChannel(userID), &c); err != nil {
		if errors.Is(err, kv.ErrNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}

	pts, ok := c[channelID]
	if !ok {
		return 0, false, nil
	}

	return pts, true, nil
}

func (s *State) SetChannelPts(userID, channelID int64, pts int) error {
	c, k := make(map[int64]int), key.StateChannel(userID)

	if err := s.Get(k, &c); err != nil {
		return err
	}
	c[channelID] = pts
	return s.Set(k, c)
}

func (s *State) ForEachChannels(userID int64, f func(channelID int64, pts int) error) error {
	c := make(map[int64]int)

	if err := s.Get(key.StateChannel(userID), &c); err != nil {
		return err
	}

	for channelID, pts := range c {
		if err := f(channelID, pts); err != nil {
			return err
		}
	}

	return nil
}
