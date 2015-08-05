package main

import (
	"errors"
)

type ChannelManager struct {
	chanMap map[string]*Channel
}

func NewChannelManager() *ChannelManager {
	this := new(ChannelManager)
	this.chanMap = make(map[string]*Channel)

	return this
}

func (this *ChannelManager) Channel(name string) (ch *Channel, err error) {
	var ok bool
	if ch, ok = this.chanMap[name]; ok {
		return ch, nil
	} else {
		return nil, errors.New("channel does not exist")
	}
}

func (this *ChannelManager) CreateChannel(name string) error {
	if _, ok := this.chanMap[name]; ok {
		return errors.New("Channel already exist")
	}

	this.chanMap[name] = NewChannel(name)
	return nil
}
