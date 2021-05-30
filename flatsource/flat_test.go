package flatsource_test

import (
	"testing"
	"unicode"

	"github.com/dangermike/runination/flatsource"
	"github.com/dangermike/runination/source"
	"github.com/dangermike/runination/testhelper"
)

func cswrapper(tables []*unicode.RangeTable) source.CharSource {
	return flatsource.New(tables)
}

func TestFlat(t *testing.T) {
	testhelper.InternalTest(t, cswrapper)
}

func BenchmarkFlatAt(b *testing.B) {
	testhelper.InternalBenchmarkAt(b, cswrapper)
}

func BenchmarkFlatCreation(b *testing.B) {
	testhelper.InternalBenchmarkCreation(b, cswrapper)
}
