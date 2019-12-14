// mautrixfs - A Matrix client as a FUSE filesystem.
// Copyright (C) 2019 Tulir Asokan
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

func (keyedState *RoomKeyedStateRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (keyedState *RoomKeyedStateRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Println("Keyed state lookup", name)

	return keyedState.NewInode(ctx, &RoomKeyedStateEvent{
		room:      keyedState.room,
		eventType: name,
		client:    keyedState.client,
	}, fs.StableAttr{Mode: syscall.S_IFDIR}), OK
}

type RoomKeyedStateEvent struct {
	fs.Inode

	room      *RoomNode
	eventType string
	client    *mautrix.Client
}

func (keyedStateEvent *RoomKeyedStateEvent) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Println("Keyed state lookup", keyedStateEvent.eventType, name)

	url := keyedStateEvent.client.BuildURL("rooms", keyedStateEvent.room.ID, "state", keyedStateEvent.eventType, name)
	data, err := keyedStateEvent.client.MakeRequest(http.MethodGet, url, nil, nil)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return keyedStateEvent.NewInode(ctx, &fs.MemRegularFile{
		Data:  data,
		Attr:  fuse.Attr{
			Mode: 0444,
		},
	}, fs.StableAttr{ Mode: syscall.S_IFREG }), OK
}
