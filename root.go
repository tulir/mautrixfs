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
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type MatrixRoot struct {
	fs.Inode

	client  *mautrix.Client
	rooms   *RoomRoot
	aliases *AliasRoot
}

func (r *MatrixRoot) OnAdd(ctx context.Context) {
	version := r.NewPersistentInode(ctx, &fs.MemRegularFile{
		Data: []byte("0.1.0"),
		Attr: fuse.Attr{Mode: 0444},
	}, fs.StableAttr{Mode: syscall.S_IFREG})
	r.AddChild("version", version, false)
	r.ForgetPersistent()

	r.aliases = &AliasRoot{client: r.client}
	r.AddChild("alias", r.NewPersistentInode(ctx, r.aliases, fs.StableAttr{Mode: syscall.S_IFDIR}), false)
	r.rooms = &RoomRoot{client: r.client}
	r.AddChild("room", r.NewPersistentInode(ctx, r.rooms, fs.StableAttr{Mode: syscall.S_IFDIR}), false)
}

func (r *MatrixRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}
