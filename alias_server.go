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
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type AliasServerRoot struct {
	fs.Inode

	server string
	client *mautrix.Client
}

var _ = (fs.NodeGetattrer)((*AliasServerRoot)(nil))
var _ = (fs.NodeLookuper)((*AliasServerRoot)(nil))
var _ = (fs.NodeUnlinker)((*AliasServerRoot)(nil))

func (aliasServer *AliasServerRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (aliasServer *AliasServerRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	resolved := aliasServer.GetChild(name)
	if resolved != nil {
		return resolved, OK
	}
	alias := fmt.Sprintf("#%s:%s", name, aliasServer.server)
	fmt.Println("Alias lookup", alias)
	data, err := aliasServer.client.ResolveAlias(alias)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return aliasServer.NewInode(ctx, &fs.MemSymlink{
		Attr: fuse.Attr{},
		Data: []byte(fmt.Sprintf("../../room/%s", data.RoomID)),
	}, fs.StableAttr{Mode: syscall.S_IFLNK}), OK
}

func (aliasServer *AliasServerRoot) Unlink(ctx context.Context, name string) syscall.Errno {
	alias := fmt.Sprintf("#%s:%s", name, aliasServer.server)
	fmt.Println("Alias unlink", alias)
	_, err := aliasServer.client.DeleteAlias(alias)
	return httpToErrno(err, true)
}
