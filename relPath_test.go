package dir

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelSym(t *testing.T) {
	testdata := []struct {
		base   string
		target string
		expVal string
		expErr error
	}{
		{
			base:   "testdata/TopA",
			target: "testdata/TopA/MiddleA/DeepA",
			expVal: "MiddleA/DeepA",
			expErr: nil,
		}, {
			base:   "testdata/TopC/BackToTopB",
			target: "testdata/TopC/BackToTopB/MiddleA",
			expVal: "MiddleA",
			expErr: nil,
		}, {
			base:   "testdata/TopC/MiddleA/BottomA/BackToTopBMiddleA/BottomC/BackToTopA",
			target: "testdata/TopC/MiddleA/BottomA/BackToTopBMiddleA/BottomC/BackToTopA/MiddleC/BottomC",
			expVal: "MiddleC/BottomC",
			expErr: nil,
		}, {
			base:   "testdata/TopB/MiddleA/BottomC/BackToTopA",
			target: "testdata/TopC/MiddleA/BottomA/BackToTopBMiddleA/BottomC/BackToTopA/MiddleC/BottomC",
			expVal: "MiddleC/BottomC",
			expErr: nil,
		}, {
			base:   "testdata/TopC/MiddleA/BottomA/BackToTopBMiddleA/BottomC/BackToTopA",
			target: "testdata/TopA/MiddleC/BottomC",
			expVal: "MiddleC/BottomC",
			expErr: nil,
		},
	}

	for id, test := range testdata {
		t.Run(fmt.Sprintf("TestRelSym #%d", id), func(t *testing.T) {
			actualVal, actualErr := RelSym(test.base, test.target)
			assert.Equal(t, test.expVal, actualVal)
			assert.Equal(t, test.expErr, actualErr)
		})
	}
}
