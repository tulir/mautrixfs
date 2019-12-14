# mautrixfs
A Matrix client as a FUSE filesystem.

This is very much work-in-progress, it barely works at all and most paths have
not been implemented.

## Design notes
* MVP: reading and writing any events and state in a room.
* After that:
  * Update caches from /sync results
  * Cache stuff on disk
  * Reading events as they come down /sync
  * Media and account data
* Slashes are bad and URL encoding isn't nice, so for room v3 event IDs, encode/decode them to show up as room v4 event IDs.
  * Room v1/v2 are only supported if users don't modify their servers to send event IDs with slashes or null characters.
* Eventually (not soon), support end-to-end encryption.

## Basic file system tree
```
/
├── version - The version of mautrixfs.
├── alias
│   └── {servername}
│       └── {localpart} - Resolves #localpart:servername and symlinks to /room/!roomid.
├── media
│   └── {server}
│       └── {id} - Data of media repo file mxc://server/id.
├── room
│   └── {roomid} - Root directory for a specific room.
│       ├── version (ro) - The version of the room.
│       ├── state
│       │   └── {event_type} (rw) - The content of the given state event type with a blank state key.
│       ├── keyed_state
│       │   └── {event_type}
│       │       └── {state_key} (rw) - The content of the given state event.
│       ├── event
│       │   └── {eventid} (ro) - The data of a specific event.
│       └── account_data
│           └── {type} (rw) - The data of a specific per-room account data event.
└── account_data
    └── {type} (rw) - The data of a specific account data event.
```
