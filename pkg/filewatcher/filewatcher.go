/*
 * Copyright (c) 2020 by The metal-stack Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package filewatcher

import (
	"context"

	"github.com/fsnotify/fsnotify"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type WatchContext interface {
	GetContext() context.Context
	logger.LogContext
}

type Event = fsnotify.Event
type Handler func(Event)

type registration struct {
	files    []string
	handlers []Handler
}

func (this *registration) copy() *registration {
	files := make([]string, len(this.files))
	copy(files, this.files)
	handlers := make([]Handler, len(this.handlers))
	copy(handlers, this.handlers)
	return &registration{
		files:    files,
		handlers: handlers,
	}
}

type FileWatcher struct {
	files         []string
	registrations []*registration
	watcher       *fsnotify.Watcher
	filehandlers  map[string][]Handler
	logger        logger.LogContext
}

func Configure() *FileWatcher {
	return &FileWatcher{}
}

func (this *FileWatcher) copy(copylast func(this *registration) bool) *FileWatcher {
	reg := len(this.registrations)
	regs := make([]*registration, reg)
	copy(regs, this.registrations)
	if len(this.registrations) > 0 && copylast(this.registrations[reg-1]) {
		regs[reg-1] = regs[reg-1].copy()
	} else {
		regs = append(regs, &registration{})
	}
	return &FileWatcher{
		registrations: regs,
	}
}

// For adds a file to watch to the actually pending watch list
func (this *FileWatcher) For(file string) *FileWatcher {
	n := this.copy(func(this *registration) bool { return len(this.handlers) != 0 })
	reg := len(n.registrations) - 1
	n.registrations[reg].files = append(n.registrations[reg].files, file)
	return n
}

// Do adds a handler to watch to the actual watch list.
// If a handler is added a new empty pending watch list is started
func (this *FileWatcher) Do(h Handler) *FileWatcher {
	n := this.copy(func(this *registration) bool { return len(this.files) != 0 })
	reg := len(n.registrations) - 1
	n.registrations[reg].handlers = append(n.registrations[reg].handlers, h)
	return n
}

// Do adds a command handler to watch to the actual watch list
// If a handler is added to a non empty pending watchlist
// this watchlist becomes the actual one and a new empty pending list is started.
func (this *FileWatcher) EnqueueCommand(c controller.Interface, cmd string) *FileWatcher {
	return this.Do(func(Event) { c.EnqueueCommand(cmd) })
}

func (this *FileWatcher) StartWith(ctx WatchContext, name string) error {
	return this.Start(ctx.GetContext(), ctx, name)
}

func (this *FileWatcher) Start(ctx context.Context, logger logger.LogContext, name string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	this.watcher = watcher
	this.logger = logger.NewContext("filewatch", name)
	this.filehandlers = map[string][]Handler{}
	done := utils.StringSet{}
	for _, r := range this.registrations {
		for _, f := range r.files {
			if f != "" {
				this.filehandlers[f] = append(this.filehandlers[f], r.handlers...)
				if !done.Contains(f) {
					done.Add(f)
					if err := this.watcher.Add(f); err != nil {
						return err
					}
				}
			}
		}
	}

	go this.watch()

	go func() {
		// Block until the stop channel is closed.
		<-ctx.Done()

		_ = this.watcher.Close()
	}()
	return nil
}

// Watch reads events from the watcher's channel and reacts to changes.
func (this *FileWatcher) watch() {
	this.logger.Infof("starting file watches %+v", this.files)
	for {
		select {
		case event, ok := <-this.watcher.Events:
			// Channel is closed.
			if !ok {
				return
			}

			this.handleEvent(event)

		case err, ok := <-this.watcher.Errors:
			// Channel is closed.
			if !ok {
				return
			}

			this.logger.Error(err, "watch error")
		}
	}
}

func (this *FileWatcher) handleEvent(event fsnotify.Event) {
	// Only care about events which may modify the contents of the file.
	if !(isWrite(event) || isRemove(event) || isCreate(event)) {
		return
	}

	this.logger.Infof("watch event %s", event)

	// If the file was removed, re-add the watch.
	if isRemove(event) {
		if err := this.watcher.Add(event.Name); err != nil {
			this.logger.Error(err, "error re-watching file")
		}
	}

	for _, h := range this.filehandlers[event.Name] {
		h(event)
	}
}

func isWrite(event fsnotify.Event) bool {
	return event.Op&fsnotify.Write == fsnotify.Write
}

func isCreate(event fsnotify.Event) bool {
	return event.Op&fsnotify.Create == fsnotify.Create
}

func isRemove(event fsnotify.Event) bool {
	return event.Op&fsnotify.Remove == fsnotify.Remove
}
