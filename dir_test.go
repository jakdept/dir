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
	"log"
	"sort"
	"testing"
)

func StaticTest(t *testing.T) {
	t.Skip()

}

func ExampleList() {
	dir, err := Watch("testdata/")
	if err != nil {
		log.Fatal(err)
	}
	list := dir.List()
	sort.Strings(list)
	bytes, _ := json.MarshalIndent(list, "", "\t")
	fmt.Println(string(bytes))
	// Output:
	// [
	// 	"/",
	// 	"/TopC",
	// 	"/topA",
	// 	"/topA/MiddleC",
	// 	"/topA/middleA",
	// 	"/topA/middleA/DeepA",
	// 	"/topA/middleA/DeepB",
	// 	"/topA/middleA/DeepC",
	// 	"/topA/middleB",
	// 	"/topA/middleB/DeepA",
	// 	"/topA/middleB/DeepB",
	// 	"/topA/middleB/DeepC",
	// 	"/topB",
	// 	"/topB/middleA",
	// 	"/topB/middleA/DeepA",
	// 	"/topB/middleA/DeepB",
	// 	"/topB/middleA/DeepC"
	// ]
}

func ExampleIn() {
	dir, err := Watch("testdata/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir.In("/topA/middleB"))
	// Output:
	// true
}
