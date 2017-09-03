package groestl256

import (
	"io"
	"testing"
)

const (
	lorem = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam laoreet justo id diam dignissim, ut feugiat dui luctus. Pellentesque vestibulum sem a est mattis aliquam. Nullam semper ut velit ut blandit. Integer vulputate rhoncus elit. Suspendisse potenti. Duis a varius arcu. Suspendisse potenti. Sed auctor quis elit vitae consequat. In eros leo, elementum id aliquet non, pretium in dolor.`
)

func TestSum(t *testing.T) {
	h := New()
	io.WriteString(h, lorem)
	t.Logf("%x\n", h.Sum(nil))
}
