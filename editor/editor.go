//editor is the place where people can come and edit their files in a safe way.
//
package editor

type Editor struct {
	Root string
}

type Tree struct {
	info     FileInfo
	children []Tree
}

func NewEditor(root string) *Editor {
	return &Editor{root}
}

func (ed *Editor) ListR(depth int) {
}
