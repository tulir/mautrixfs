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
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type RoomNode struct {
	fs.Inode
	mautrix.Room

	Version string
	client  *mautrix.Client
}

func (room *RoomNode) OnAdd(ctx context.Context) {
	version := room.NewPersistentInode(ctx, &fs.MemRegularFile{
		Data: []byte(room.Version),
		Attr: fuse.Attr{Mode: 0444},
	}, fs.StableAttr{Mode: syscall.S_IFREG})
	room.AddChild("version", version, false)
	eventsInode := room.NewPersistentInode(ctx, &RoomEventRoot{room: room, client: room.client}, fs.StableAttr{Mode: syscall.S_IFDIR})
	room.AddChild("event", eventsInode, false)
	stateInode := room.NewPersistentInode(ctx, &RoomStateRoot{room: room, client: room.client}, fs.StableAttr{Mode: syscall.S_IFDIR})
	room.AddChild("state", stateInode, false)
	keyedStateInode := room.NewPersistentInode(ctx, &RoomKeyedStateRoot{room: room, client: room.client}, fs.StableAttr{Mode: syscall.S_IFDIR})
	room.AddChild("keyed_state", keyedStateInode, false)
}

func (room *RoomNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}
