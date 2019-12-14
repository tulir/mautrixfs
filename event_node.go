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

type RoomEventNode struct {
	fs.Inode

	room   *RoomNode
	id     string
	data   []byte
	client *mautrix.Client
}

var _ = (fs.NodeGetattrer)((*RoomEventNode)(nil))
var _ = (fs.NodeOpener)((*RoomEventNode)(nil))
var _ = (fs.NodeReader)((*RoomEventNode)(nil))

func (event *RoomEventNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0444
	out.Size = uint64(len(event.data))
	return OK
}

func (event *RoomEventNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, fuse.FOPEN_KEEP_CACHE, OK
}

func (event *RoomEventNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	end := int(off) + len(dest)
	if end > len(event.data) {
		end = len(event.data)
	}
	return fuse.ReadResultData(event.data[off:end]), OK
}
