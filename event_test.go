package ev_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

type mockFileInfo struct{ name string }

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() any           { return nil }

func TestOp_String(t *testing.T) {
	tests := []struct {
		op       ev.Op
		expected string
	}{
		{ev.CHMOD, "CHMOD"},
		{ev.CREATE, "CREATE"},
		{ev.REMOVE, "REMOVE"},
		{ev.RENAME, "RENAME"},
		{ev.WRITE, "WRITE"},
		{ev.Op(0), "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if got := test.op.String(); got != test.expected {
				t.Errorf("Op.String() = %q, expected %q", got, test.expected)
			}
		})
	}
}

func TestEvent_Op(t *testing.T) {
	tests := []struct {
		name string
		op   ev.Op
	}{
		{"chmod", ev.CHMOD},
		{"create", ev.CREATE},
		{"remove", ev.REMOVE},
		{"rename", ev.RENAME},
		{"write", ev.WRITE},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := ev.NewEvent(test.op, "", nil)
			if got := e.Op(); got != test.op {
				t.Errorf("Event.Op() = %v, expected %v", got, test.op)
			}
		})
	}
}

func TestEvent_Path(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"absolute path", "/home/user/file.go"},
		{"relative path", "cmd/main.go"},
		{"empty path", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := ev.NewEvent(ev.WRITE, test.path, nil)
			if got := e.Path(); got != test.path {
				t.Errorf("Event.Path() = %q, expected %q", got, test.path)
			}
		})
	}
}

func TestEvent_Info(t *testing.T) {
	info := mockFileInfo{name: "file.go"}
	tests := []struct {
		name string
		info fs.FileInfo
	}{
		{"non-nil info", info},
		{"nil info (manual trigger)", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := ev.NewEvent(ev.CREATE, "/some/path", test.info)
			if got := e.Info(); got != test.info {
				t.Errorf("Event.Info() = %v, expected %v", got, test.info)
			}
		})
	}
}

func TestEvent_Has(t *testing.T) {
	tests := []struct {
		name     string
		op       ev.Op
		check    ev.Op
		expected bool
	}{
		{"exact match write", ev.WRITE, ev.WRITE, true},
		{"exact match create", ev.CREATE, ev.CREATE, true},
		{"no match", ev.WRITE, ev.CREATE, false},
		{"combined op has write", ev.WRITE | ev.CREATE, ev.WRITE, true},
		{"combined op has create", ev.WRITE | ev.CREATE, ev.CREATE, true},
		{"combined op missing remove", ev.WRITE | ev.CREATE, ev.REMOVE, false},
		{"zero op", ev.Op(0), ev.WRITE, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := ev.NewEvent(test.op, "", nil)
			if got := e.Has(test.check); got != test.expected {
				t.Errorf("Event.Has(%v) = %v, expected %v", test.check, got, test.expected)
			}
		})
	}
}
