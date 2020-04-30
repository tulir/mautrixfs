// mautrixfs - A Matrix client as a FUSE filesystem.
// Copyright (C) 2020 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"net/http"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type RoomKeyedStateRoot struct {
	fs.Inode

	room   *RoomNode
	client *mautrix.Client
}

var _ fs.NodeGetattrer = (*RoomKeyedStateRoot)(nil)
var _ fs.NodeLookuper = (*RoomKeyedStateRoot)(nil)
var _ fs.NodeUnlinker = (*RoomKeyedStateRoot)(nil)

func (keyedState *RoomKeyedStateRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (keyedState *RoomKeyedStateRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	resolved := keyedState.GetChild(name)
	if resolved != nil {
		return resolved, OK
	}
	fmt.Println("Keyed state lookup", name)

	return keyedState.NewInode(ctx, &RoomKeyedStateEvent{
		room:      keyedState.room,
		eventType: name,
		client:    keyedState.client,
	}, fs.StableAttr{Mode: syscall.S_IFDIR}), OK
}

func (keyedState *RoomKeyedStateRoot) Unlink(ctx context.Context, name string) syscall.Errno {
	return syscall.EROFS
}

type RoomKeyedStateEvent struct {
	fs.Inode

	room      *RoomNode
	eventType string
	client    *mautrix.Client
}

var _ fs.NodeGetattrer = (*RoomKeyedStateEvent)(nil)
var _ fs.NodeLookuper = (*RoomKeyedStateEvent)(nil)
var _ fs.NodeUnlinker = (*RoomKeyedStateEvent)(nil)

func (keyedStateEvent *RoomKeyedStateEvent) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (keyedStateEvent *RoomKeyedStateEvent) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	resolved := keyedStateEvent.GetChild(name)
	if resolved != nil {
		return resolved, OK
	}
	fmt.Println("Keyed state lookup", keyedStateEvent.eventType, name)

	url := keyedStateEvent.client.BuildURL("rooms", keyedStateEvent.room.ID, "state", keyedStateEvent.eventType, name)
	data, err := keyedStateEvent.client.MakeRequest(http.MethodGet, url, nil, nil)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return keyedStateEvent.NewInode(ctx, &StateEventNode{
		client:    keyedStateEvent.client,
		room:      keyedStateEvent.room,
		eventType: keyedStateEvent.eventType,
		stateKey:  name,
		data:      data,
	}, fs.StableAttr{Mode: syscall.S_IFREG}), OK
}

func (keyedStateEvent *RoomKeyedStateEvent) Unlink(ctx context.Context, name string) syscall.Errno {
	return syscall.EROFS
}
