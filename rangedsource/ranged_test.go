package rangedsource_test

import (
	"testing"
	"unicode"

	"github.com/dangermike/runination/rangedsource"
	"github.com/dangermike/runination/source"
	"github.com/dangermike/runination/testhelper"
)

func cswrapper(tables []*unicode.RangeTable) source.CharSource {
	return rangedsource.New(tables)
}

func TestRanged(t *testing.T) {
	testhelper.InternalTest(t, cswrapper)
}

func BenchmarkRangedAt(b *testing.B) {
	testhelper.InternalBenchmarkAt(b, cswrapper)
}

func BenchmarkRangedCreation(b *testing.B) {
	testhelper.InternalBenchmarkCreation(b, cswrapper)
}
