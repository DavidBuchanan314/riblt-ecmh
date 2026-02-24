package riblt

import "github.com/gtank/ristretto255"

// Sketch is a prefix of the coded symbol sequence for a set of source symbols.
// When generating a prefix of predetermined length, compared to generating the
// prefix incrementally using an Encoder, it is more efficient to use Sketch.
// Sketch also allows inserting or deleting source symbols from the set after
// it has been created.
type Sketch[T Symbol[T]] []CodedSymbol[T]

// NewSketch creates a Sketch of the given size with all checksums initialized
// to the identity element.
func NewSketch[T Symbol[T]](size int) Sketch[T] {
	s := make(Sketch[T], size)
	for i := range s {
		s[i].Checksum = ristretto255.NewIdentityElement()
	}
	return s
}

// AddHashedSymbol inserts source symbol t to the set of which s is a sketch.
func (s Sketch[T]) AddHashedSymbol(t HashedSymbol[T]) {
	m := randomMapping{t.Hash, 0}
	for m.lastIndex < len(s) {
		idx := m.lastIndex
		s[idx].Symbol = s[idx].Symbol.XOR(t.Symbol)
		s[idx].Count += 1
		s[idx].Checksum = new(ristretto255.Element).Add(s[idx].Checksum, t.Checksum)
		m.nextIndex()
	}
}

// RemoveHashedSymbol deletes source symbol t from the set of which s is a
// sketch.
func (s Sketch[T]) RemoveHashedSymbol(t HashedSymbol[T]) {
	m := randomMapping{t.Hash, 0}
	for m.lastIndex < len(s) {
		idx := m.lastIndex
		s[idx].Symbol = s[idx].Symbol.XOR(t.Symbol)
		s[idx].Count -= 1
		s[idx].Checksum = new(ristretto255.Element).Subtract(s[idx].Checksum, t.Checksum)
		m.nextIndex()
	}
}

// AddSymbol inserts source symbol t to the set of which s is a sketch.
func (s Sketch[T]) AddSymbol(t T) {
	s.AddHashedSymbol(newHashedSymbol(t))
}

// RemoveSymbol deletes source symbol t from the set of which s is a sketch.
func (s Sketch[T]) RemoveSymbol(t T) {
	s.RemoveHashedSymbol(newHashedSymbol(t))
}

// Subtract subtracts s2 from s by modifying s in place. s and s2 must be of
// equal length. If s is a sketch of set S and s2 is a sketch of set S2, then
// the result is a sketch of the symmetric difference between S and S2.
func (s Sketch[T]) Subtract(s2 Sketch[T]) {
	if len(s) != len(s2) {
		panic("subtracting sketches of different sizes")
	}

	for i := range s {
		s[i].Symbol = s[i].Symbol.XOR(s2[i].Symbol)
		s[i].Count = s[i].Count - s2[i].Count
		s[i].Checksum = new(ristretto255.Element).Subtract(s[i].Checksum, s2[i].Checksum)
	}
}

// Decode tries to decode s, where s can be one of the following
//  1. A sketch of set S.
//  2. Content of s after calling s.Subtract(s2), where s is a sketch of set
//     S, and s2 is a sketch of set S2.
//
// When successful, indicated by succ being true, fwd contains all source
// symbols in S in case 1, or S \ S2 in case 2 (\ is the set subtraction
// operation). rev is empty in case 1, or S2 \ S in case 2.
func (s Sketch[T]) Decode() (fwd []HashedSymbol[T], rev []HashedSymbol[T], succ bool) {
	dec := Decoder[T]{}
	for _, c := range s {
		dec.AddCodedSymbol(c)
	}
	dec.TryDecode()
	return dec.Remote(), dec.Local(), dec.Decoded()
}
