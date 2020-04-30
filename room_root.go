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
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type RoomRoot struct {
	fs.Inode

	client *mautrix.Client
	rooms  map[string]*RoomNode
}

func (r *RoomRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (r *RoomRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if child := r.GetChild(name); child != nil {
		return child, OK
	}
	var content = make(map[string]interface{})
	roomID := id.RoomID(name)
	err := r.client.StateEvent(roomID, event.StateCreate, "", &content)
	if err != nil {
		return nil, syscall.ENOENT
	}
	version, ok := content["room_version"].(string)
	if !ok {
		version = "1"
	}

	roomNode := &RoomNode{
		Room: mautrix.Room{
			ID:    roomID,
			State: make(map[event.Type]map[string]*event.Event),
		},
		Version: version,
		client:  r.client,
	}
	inode := r.NewInode(ctx, roomNode, fs.StableAttr{Mode: syscall.S_IFDIR})
	return inode, OK
}

type RoomListStream struct {
	next int
	data []id.RoomID
}

func (rls *RoomListStream) HasNext() bool {
	return rls.next < len(rls.data)
}

func (rls *RoomListStream) Next() (fuse.DirEntry, syscall.Errno) {
	rls.next += 1
	return fuse.DirEntry{
		Mode: 0555,
		Name: string(rls.data[rls.next-1]),
	}, OK
}

func (rls *RoomListStream) Close() {}

func (r *RoomRoot) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	fmt.Println("Readdir rooms")
	resp, err := r.client.JoinedRooms()
	if err != nil {
		return nil, syscall.EIO
	}

	return &RoomListStream{data: resp.JoinedRooms}, OK
}
