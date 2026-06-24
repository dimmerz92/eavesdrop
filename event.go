package ev

import (
	"io/fs"

	"github.com/fsnotify/fsnotify"
)

type Op uint32

const (
	CHMOD  Op = Op(fsnotify.Chmod)
	CREATE Op = Op(fsnotify.Create)
	REMOVE Op = Op(fsnotify.Remove)
	RENAME Op = Op(fsnotify.Rename)
	WRITE  Op = Op(fsnotify.Write)
)

// String returns the file operation as a string.
func (o Op) String() string {
	var op string
	switch o {
	case CHMOD:
		op = "CHMOD"
	case CREATE:
		op = "CREATE"
	case REMOVE:
		op = "REMOVE"
	case RENAME:
		op = "RENAME"
	case WRITE:
		op = "WRITE"
	}
	return op
}

// Event represents a file system change notification.
type Event struct {
	op   Op
	path string
	info fs.FileInfo
}

// NewEvent constructs an Event with the given operation, path, and file info.
func NewEvent(op Op, path string, info fs.FileInfo) Event {
	return Event{op: op, path: path, info: info}
}

// Has returns true if the event has the given operation, otherwise false.
func (e Event) Has(op Op) bool { return (e.op & op) > 0 }

// Op returns the file operation that emitted the event.
func (e Event) Op() Op { return e.op }

// Path returns the absolute file path of the affected file or directory.
func (e Event) Path() string { return e.path }

// Info returns the file info at the time the event was emitted. May be nil if
// the event was manually triggered.
func (e Event) Info() fs.FileInfo { return e.info }
