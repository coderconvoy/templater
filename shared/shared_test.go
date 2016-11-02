package shared

import (
	"github.com/russross/blackfriday"
	"testing"
)

func Test_HeadedMD(t *testing.T) {
	b := []byte("Hello\n#\nGoodbye\n====\nI did a poo")
	m := ParseHeadedMD(b)
	if m["contents"] != string(blackfriday.MarkdownCommon([]byte("Goodbye\n====\nI did a poo"))) {
		t.Fail()
		t.Log("Contents:\n", m["contents"])
	}
	if m["head"] != "Hello" {
		t.Fail()
		t.Log("Head:", m["head"])
	}
}
