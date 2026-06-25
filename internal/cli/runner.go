package cli

import (
	"fmt"
	"sync"

	ev "github.com/dimmerz92/eavesdrop"
	"github.com/fatih/color"
)

func NewShellRunner(shell *ev.Shell, name string, mu *sync.Mutex, tasks []string, service string) func(ev.Event) {
	return func(event ev.Event) {
		mu.Lock()
		defer mu.Unlock()

		err := shell.KillProcessGroup()
		if err != nil {
			color.Red("%s: failed to kill previous service: %v", name, err)
		}

		for _, task := range tasks {
			fmt.Printf("%s: running task: %s\n", color.CyanString(name), task)
			err := shell.ExecAndWait(task)
			if err != nil {
				color.Red("%s: failed to run task: %v", name, err)
			}
		}

		if service != "" {
			fmt.Printf("%s: running service: %s\n", color.BlueString(name), service)
			err := shell.ExecAndReturn(service)
			if err != nil {
				color.Red("%s: failed to run service: %v", name, err)
			}
		}
	}
}
