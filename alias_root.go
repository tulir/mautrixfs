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
	"regexp"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type AliasRoot struct {
	fs.Inode

	client *mautrix.Client
}

var _ fs.NodeGetattrer = (*AliasRoot)(nil)
var _ fs.NodeLookuper = (*AliasRoot)(nil)

var ServerNameRegex = regexp.MustCompile(`(?:(?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})|(?:\[(?:[0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}])|(?:[\w-.]{1,255}))(?::\d{1,5})?`)

func (alias *AliasRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (alias *AliasRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	resolved := alias.GetChild(name)
	if resolved != nil {
		return resolved, OK
	} else if !ServerNameRegex.MatchString(name) {
		return nil, syscall.ENOENT
	}
	return alias.NewInode(ctx, &AliasServerRoot{
		client: alias.client,
		server: name,
	}, fs.StableAttr{Mode: syscall.S_IFDIR}), OK
}
