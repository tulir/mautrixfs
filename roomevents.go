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
	"net/http"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type RoomEventRoot struct {
	fs.Inode

	room *RoomNode
	client *mautrix.Client
}

func (events *RoomEventRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (events *RoomEventRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if events.room.Version == "3" {
		name = strings.ReplaceAll(name,"-", "+")
		name = strings.ReplaceAll(name,"_", "/")
	}
	url := events.client.BuildURL("rooms", events.room.ID, "event", name)
	data, err := events.client.MakeRequest(http.MethodGet, url, nil, nil)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return events.NewInode(ctx, &fs.MemRegularFile{
		Data:  data,
		Attr:  fuse.Attr{
			Mode: 0444,
		},
	}, fs.StableAttr{ Mode: syscall.S_IFREG }), OK
}
