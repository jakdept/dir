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

// Package dir provides a mechanism to track all directories under a location.
// It can be queried at any time for a list of all directories uderneath, and
// any directory can be checked for within that location at any time. Paths
// within are treated as if chrooted - absolute path is measured from the
// tracking point.
//
// There will likely be a delay of a few seconds before new directories are
// picked up, but the delay is platform specific.
//
// When done, it should be closed.
//
// Watch should be used to start tracking a directory, as startup is needed.
package dir

import (
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/rjeczalik/notify"
)

// Dir is a type used to track folders under a specific location. It will scan
// that location recursively upon startup, and watch for further directory
// creation, removal, renaming, and the like.
type Dir struct {
	dirs     map[string]interface{}
	basepath string
	lock     sync.RWMutex
	isClosed bool
	updates  chan notify.EventInfo
}

// In allows to see if any given path within the trackced paths is present.
func (d *Dir) In(s string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if d.isClosed {
		return false
	}

	_, isIn := d.dirs[filepath.Clean(s)]
	return isIn
}

// List will return all directories within the location as currently tracked.
func (d *Dir) List() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if d.isClosed {
		return []string{}
	}

	var dirs []string
	for each := range d.dirs {
		dirs = append(dirs, each)
	}
	return dirs
}

func (d *Dir) walkFunc() filepath.WalkFunc {
	return func(loc string, info os.FileInfo, err error) error {
		// might not handle symlinks - remember to check for that
		if !info.IsDir() {
			return nil
		}

		d.lock.Lock()
		defer d.lock.Unlock()
		d.dirs[d.makePath(loc)] = true
		return nil
	}
}

func (d *Dir) makePath(p string) string {
	p, _ = filepath.Abs(p)
	p, _ = filepath.EvalSymlinks(p)
	p, _ = filepath.Rel(d.basepath, p)
	return path.Clean("/" + p)
}

func (d *Dir) updateDir(e notify.EventInfo) {
	d.lock.Lock()
	defer d.lock.Unlock()

	// bail out if it's not a directory
	info, _ := os.Stat(e.Path())
	if !info.IsDir() {
		return
	}

	switch e.Event() {
	case notify.Create:
		d.dirs[d.makePath(e.Path())] = true
	case notify.Rename:
		// apparently a rename operation fires off two events?
		// https://github.com/rjeczalik/notify/issues/78
		_, err := os.Stat(e.Path())
		if err != nil {
			delete(d.dirs, d.makePath(e.Path()))
		} else {
			d.dirs[d.makePath(e.Path())] = true
		}
	case notify.Remove:
		delete(d.dirs, d.makePath(e.Path()))
	}
}

func (d *Dir) processEvents() {
	func() {
		for e := range d.updates {
			go d.updateDir(e)
		}
	}()
}

// Close stops tracking the directory structure and closes it.
func (d *Dir) Close() {
	d.lock.Lock()
	defer d.lock.Unlock()
	notify.Stop(d.updates)
	d.isClosed = true
	// close(d.updates)
}

// Watch is used to start watching a given location for updates. Once run, the
// returned Dir can be queried immediately for current directory contents, and
// will pick up changes after a short delay.
func Watch(path string) (*Dir, error) {
	var d Dir

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, err
	}

	path, _ = filepath.Abs(path)
	d.basepath, _ = filepath.EvalSymlinks(path)
	d.updates = make(chan notify.EventInfo, 10)
	d.dirs = make(map[string]interface{})
	go d.processEvents()

	if err = notify.Watch(path, d.updates, notify.All); err != nil {
		return nil, err
	}

	err = filepath.Walk(path, d.walkFunc())
	if err != nil {
		return nil, err
	}
	return &d, nil
}
