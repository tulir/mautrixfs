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

type RoomStateRoot struct {
	fs.Inode

	room   *RoomNode
	client *mautrix.Client
}

var _ fs.NodeGetattrer = (*RoomStateRoot)(nil)
var _ fs.NodeLookuper = (*RoomStateRoot)(nil)
var _ fs.NodeUnlinker = (*RoomStateRoot)(nil)

func (state *RoomStateRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (state *RoomStateRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	resolved := state.GetChild(name)
	if resolved != nil {
		return resolved, OK
	}
	fmt.Println("State lookup", name)

	url := state.client.BuildURL("rooms", state.room.ID, "state", name)
	data, err := state.client.MakeRequest(http.MethodGet, url, nil, nil)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return state.NewInode(ctx, &StateEventNode{
		client:    state.client,
		room:      state.room,
		eventType: name,
		stateKey:  "",
		data:      data,
	}, fs.StableAttr{Mode: syscall.S_IFREG}), OK
}

func (state *RoomStateRoot) Unlink(ctx context.Context, name string) syscall.Errno {
	return syscall.EROFS
}
