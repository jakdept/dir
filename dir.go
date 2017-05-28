// MIT License
//
// Copyright (c) 2017 Jack Hayhurst
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package dir

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/rjeczalik/notify"
)

type Dir struct {
	dirs    map[string]interface{}
	lock    sync.RWMutex
	updates chan notify.EventInfo
}

func (d *Dir) In(s string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()

	_, isIn := d.dirs[s]
	return isIn
}

func (d *Dir) List() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	var dirs []string
	for each := range d.dirs {
		dirs = append(dirs, each)
	}
	return dirs
}

func (d *Dir) WalkFunc(path string, info os.FileInfo, err error) error {
	// might not handle symlinks - remember to check for that
	if !info.IsDir() {
		return nil
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	// maybe doesn't strip the prefix from the path?
	d.dirs[path] = true
	return nil
}

func (d *Dir) updateDir(e notify.EventInfo) {
	d.lock.Lock()
	defer d.lock.Unlock()

	switch e.Event() {
	case notify.Create:
		d.dirs[e.Path()] = true
	case notify.Rename:
		// apparently a rename operation fires off two events?
		// https://github.com/rjeczalik/notify/issues/78
		_, err := os.Stat(e.Path())
		if err != nil {
			delete(d.dirs, e.Path())
		} else {
			d.dirs[e.Path()] = true
		}
	case notify.Remove:
		delete(d.dirs, e.Path())
	}
}

func (d *Dir) processEvents() {
	func() {
		for e := range d.updates {
			go d.updateDir(e)
		}
	}()
}

func (d *Dir) Close() {
	notify.Stop(d.updates)
	close(d.updates)
}

func Watch(path string) (*Dir, error) {
	var d Dir

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, err
	}

	d.updates = make(chan notify.EventInfo, 100)
	d.processEvents()
	defer d.Close()

	err = notify.Watch(path, d.updates, notify.FSEventsIsDir)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(path, d.WalkFunc)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
