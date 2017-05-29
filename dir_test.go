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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

// This is a basic example showing how to list directories under a given location.
func ExampleDir_List() {
	// Open the location - in this case, the test data.
	dir, err := Watch("testdata/")
	if err != nil {
		log.Fatal(err)
	}
	// Close it when done
	defer dir.Close()

	// List the contents. The rest is consistent formatting.
	list := dir.List()
	sort.Strings(list)
	bytes, _ := json.MarshalIndent(list, "", "\t")
	fmt.Println(string(bytes))
}

func TestList(t *testing.T) {
	// Open the location - in this case, the test data.
	dir, err := Watch("testdata/")
	if err != nil {
		log.Fatal(err)
	}
	// Close it when done
	defer dir.Close()

	// List the contents. The rest is consistent formatting.
	list := dir.List()
	sort.Strings(list)
	bytes, _ := json.MarshalIndent(list, "", "\t")
	goldie.Assert(t, "TestStatic", bytes)
}

// This is an example showing how to test if something is present under the location.
func ExampleDir_In() {
	// Open the location, and close it when done with this.
	dir, err := Watch("testdata/")
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	// Test to see if topA/middleB is a sub directory of testdata/
	fmt.Println(dir.In("/topA/middleB"))
	// Output:
	// true
}

func TestLive(t *testing.T) {
	folders := []string{
		"/apple",
		"/banana",
		"/carrot",
		"/carrot/celery",
		"/dog/dolphin",
	}

	notFolders := []string{
		"/apricot",
		"/bubble",
		"celery",
	}

	base, err := ioutil.TempDir("", "jakdept.dir-")
	if err != nil {
		t.Fatalf("failed to create tempdir - %v", err)
	}

	dir, err := Watch(base)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	for _, each := range folders {
		err = os.MkdirAll(filepath.Join(base, each), 0750)
		if err != nil {
			t.Fatalf("failed to create directory - %v", err)
		}
	}

	f, err := os.Create(filepath.Join(base, "junkfile"))
	if err != nil {
		t.Fatalf("failed creating a file: %v", err)
	}
	f.Close()

	for !dir.In(folders[0]) {
		time.Sleep(5 * time.Second)
	}

	assert.True(t, dir.In(folders[0]))
	// for _, each := range folders {
	// 	assert.True(t, dir.In(each))
	// }

	for _, each := range notFolders {
		assert.False(t, dir.In(each))
	}

	// cannot verify list as not all directories will show up right away
}
