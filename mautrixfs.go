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
	"log"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"maunium.net/go/mautrix"
)

const OK = 0

func main() {
	mautrix.DisableFancyEventParsing = true
	client, err := mautrix.NewClient(os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		log.Fatalf("Failed to create client: %v\n", err)
	}
	opts := &fs.Options{MountOptions: fuse.MountOptions{Debug: true}}
	server, err := fs.Mount(os.Args[1], &MatrixRoot{
		client: client,
	}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	server.Wait()
}
