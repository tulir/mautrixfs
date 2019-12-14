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

type RoomStateRoot struct {
	fs.Inode

	room *RoomNode
	client *mautrix.Client
}

func (state *RoomStateRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (state *RoomStateRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Println("State lookup", name)
	url := state.client.BuildURL("rooms", state.room.ID, "state", name)
	data, err := state.client.MakeRequest(http.MethodGet, url, nil, nil)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return state.NewInode(ctx, &fs.MemRegularFile{
		Data:  data,
		Attr:  fuse.Attr{
			Mode: 0444,
		},
	}, fs.StableAttr{ Mode: syscall.S_IFREG }), OK
}
