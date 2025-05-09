package notify_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/notify"
)

func TestNewWatcher(t *testing.T) {
	watcher := notify.NewWatcher(config.DefaultConfig(""))

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

	if !reflect.DeepEqual(watcher.IgnoreDirs, expected["ignored_dirs"]) {
		t.Errorf("ignored_dirs expected %+v\ngot %+v", expected["ignored_dirs"], watcher.IgnoreDirs)
	}

	if !reflect.DeepEqual(watcher.IgnoreFiles, expected["ignored_files"]) {
		t.Errorf("ignored_files expected %+v\ngot %+v", expected["ignored_files"], watcher.IgnoreFiles)
	}

	if !reflect.DeepEqual(watcher.IgnoreRegex, expected["ignored_regex"]) {
		t.Errorf("ignored_regex expected %+v\ngot %+v", expected["ignored_regex"], watcher.IgnoreRegex)
	}

	if !reflect.DeepEqual(watcher.WatchFiles, expected["watched_files"]) {
		t.Errorf("watched_files expected %+v\ngot %+v", expected["watched_files"], watcher.WatchFiles)
	}

	if !reflect.DeepEqual(watcher.WatchExts, expected["watched_exts"]) {
		t.Errorf("watched_exts expected %+v\ngot %+v", expected["watched_exts"], watcher.WatchExts)
	}
}
