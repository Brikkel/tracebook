package diff

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

func GetDiff(before, after string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(before, after, false)
	return dmp.DiffPrettyText(diffs)
}
