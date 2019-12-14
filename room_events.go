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
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

type RoomEventRoot struct {
	fs.Inode

	room   *RoomNode
	client *mautrix.Client
}

var _ = (fs.NodeGetattrer)((*RoomEventRoot)(nil))
var _ = (fs.NodeLookuper)((*RoomEventRoot)(nil))
var _ = (fs.NodeUnlinker)((*RoomEventRoot)(nil))

func (events *RoomEventRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0555
	return OK
}

func (events *RoomEventRoot) mutateEventID(eventID string) string {
	if events.room.Version == "3" {
		eventID = strings.ReplaceAll(eventID, "-", "+")
		eventID = strings.ReplaceAll(eventID, "_", "/")
	}
	return eventID
}

func (events *RoomEventRoot) Unlink(ctx context.Context, name string) syscall.Errno {
	eventID := events.mutateEventID(name)
	fmt.Println("Event unlink", eventID)
	_, err := events.client.RedactEvent(events.room.ID, eventID)
	return httpToErrno(err, true)
}

func (events *RoomEventRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	eventID := events.mutateEventID(name)
	fmt.Println("Event lookup", eventID)

	url := events.client.BuildURL("rooms", events.room.ID, "event", eventID)
	data, err := events.client.MakeRequest(http.MethodGet, url, nil, nil)
	if err != nil || data == nil {
		return nil, syscall.ENOENT
	}

	return events.NewInode(ctx, &RoomEventNode{
		room: events.room,
		id:   eventID,
		data: data,
	}, fs.StableAttr{Mode: syscall.S_IFREG}), OK
}
