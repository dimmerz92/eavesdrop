package notify_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/notify"
	"github.com/fsnotify/fsnotify"
)

func TestNewNotifier(t *testing.T) {
	notifier := notify.NewNotifier(config.DefaultConfig(""))

	expected := map[string]any{
		"ignored_dirs": map[string]struct{}{
			"assets": {}, "data": {}, "node_modules": {}, "testdata": {}, "tmp": {}, "vendor": {},
		},
		"ignored_files": map[string]struct{}{},
		"ignored_regex": []*regexp.Regexp{
			regexp.MustCompile(`^\.{1}.*$`), regexp.MustCompile(`^.*_templ\.go$`), regexp.MustCompile(`^.*_test\.go$`),
		},
		"watched_files": map[string]struct{}{},
		"watched_exts":  map[string]struct{}{".go": {}, ".html": {}, ".templ": {}, ".tmpl": {}, ".tpl": {}},
	}

	if !reflect.DeepEqual(notifier.IgnoreDirs, expected["ignored_dirs"]) {
		t.Errorf("ignored_dirs expected %+v\ngot %+v", expected["ignored_dirs"], notifier.IgnoreDirs)
	}

	if !reflect.DeepEqual(notifier.IgnoreFiles, expected["ignored_files"]) {
		t.Errorf("ignored_files expected %+v\ngot %+v", expected["ignored_files"], notifier.IgnoreFiles)
	}

	if !reflect.DeepEqual(notifier.IgnoreRegex, expected["ignored_regex"]) {
		t.Errorf("ignored_regex expected %+v\ngot %+v", expected["ignored_regex"], notifier.IgnoreRegex)
	}

	if !reflect.DeepEqual(notifier.WatchFiles, expected["watched_files"]) {
		t.Errorf("watched_files expected %+v\ngot %+v", expected["watched_files"], notifier.WatchFiles)
	}

	if !reflect.DeepEqual(notifier.WatchExts, expected["watched_exts"]) {
		t.Errorf("watched_exts expected %+v\ngot %+v", expected["watched_exts"], notifier.WatchExts)
	}
}

func TestNotifier_ShouldIgnore(t *testing.T) {
	notifier := notify.NewNotifier(config.DefaultConfig(""))

	t.Run("test directories", func(t *testing.T) {
		tests := map[string]bool{
			// named directories
			"assets": true, "data": true, "node_modules": true, "testdata": true, "tmp": true, "vendor": true,
			// regex patterned
			".git": true, ".DS_STORE": true,
			// should not ignore
			"internal": false, "cmd": false, "sql": false,
		}

		for k, v := range tests {
			if ok := notifier.ShouldIgnore(fsnotify.Event{Name: k}, true); ok != v {
				t.Errorf("%s expected %t got %t", k, v, ok)
			}
		}
	})

	t.Run("test files", func(t *testing.T) {
		notifier.IgnoreFiles = map[string]struct{}{"ignored.go": {}}
		notifier.WatchFiles = map[string]struct{}{"test.txt": {}, ".config": {}}

		tests := map[string]bool{
			// named directories
			"ignored.go": true,
			// regex patterned
			".env": true, "page_templ.go": true, "test_test.go": true,
			// should not ignore
			"test.txt": false, ".config": false,
		}

		for k, v := range tests {
			if ok := notifier.ShouldIgnore(fsnotify.Event{Name: k}, false); ok != v {
				t.Errorf("%s expected %t got %t", k, v, ok)
			}
		}
	})
}

/*
func TestShouldIgnoreFile(t *testing.T) {
}
*/
