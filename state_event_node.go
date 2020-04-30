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
	"encoding/json"
	"fmt"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type StateEventNode struct {
	fs.Inode

	room      *RoomNode
	eventType string
	stateKey  string
	data      []byte

	writeMutex sync.Mutex
	writeData  []byte

	client *mautrix.Client
}

var _ fs.NodeGetattrer = (*StateEventNode)(nil)
var _ fs.NodeOpener = (*StateEventNode)(nil)
var _ fs.NodeReader = (*StateEventNode)(nil)
var _ fs.NodeWriter = (*StateEventNode)(nil)
var _ fs.NodeFlusher = (*StateEventNode)(nil)

func (stateEvent *StateEventNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0644
	out.Size = uint64(len(stateEvent.data))
	return OK
}

func (stateEvent *StateEventNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, fuse.FOPEN_KEEP_CACHE, OK
}

func (stateEvent *StateEventNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	end := int(off) + len(dest)
	if end > len(stateEvent.data) {
		end = len(stateEvent.data)
	}
	return fuse.ReadResultData(stateEvent.data[off:end]), OK
}

func (stateEvent *StateEventNode) Write(ctx context.Context, fh fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	stateEvent.writeMutex.Lock()
	defer stateEvent.writeMutex.Unlock()
	end := int64(len(data)) + off
	if int64(len(stateEvent.writeData)) < end {
		n := make([]byte, end)
		copy(n, stateEvent.writeData)
		stateEvent.writeData = n
	}

	copy(stateEvent.writeData[off:off+int64(len(data))], data)

	return uint32(len(data)), OK
}

func (stateEvent *StateEventNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	stateEvent.writeMutex.Lock()
	defer stateEvent.writeMutex.Unlock()
	if sz, ok := in.GetSize(); ok {
		stateEvent.writeData = stateEvent.writeData[:sz]
	}
	out.Mode = 0644
	out.Size = uint64(len(stateEvent.writeData))
	return OK
}

func (stateEvent *StateEventNode) Flush(ctx context.Context, fh fs.FileHandle) syscall.Errno {
	if len(stateEvent.writeData) == 0 {
		return OK
	}
	var data map[string]interface{}
	err := json.Unmarshal(stateEvent.writeData, &data)
	if err != nil {
		fmt.Println(err)
		return syscall.EBADMSG
	}
	_, err = stateEvent.client.SendStateEvent(stateEvent.room.ID,
		event.Type{Type: stateEvent.eventType, Class: event.StateEventType}, stateEvent.stateKey, data)
	if err == nil {
		stateEvent.data, _ = json.Marshal(data)
	}
	return httpToErrno(err, false)
}
