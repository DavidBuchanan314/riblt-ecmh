package riblt

import (
	"testing"
)

func BenchmarkMapping(b *testing.B) {
	m := newRandomMapping([32]byte{1, 2, 3, 4, 5, 6, 7, 8})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.nextIndex()
	}
}
