package lzma

import (
	"io"
	"testing"
)

func TestHTable(t *testing.T) {
	var ht htable
	for u := uint32(0); u < 512; u++ {
		ht.put(u, pointer(int(u)))
	}
	if len(ht.table) != 1024 {
		t.Fatalf("len(ht.table) is %d; want %d", len(ht.table), 1024)
	}
	for u := uint32(0); u < 512; u++ {
		p, ok := ht.get(u)
		if !ok {
			t.Fatalf("get(%d) can't find a value", u)
		}
		w := pointer(int(u))
		if p != w {
			t.Fatalf("get(%d) returned %d; want %d", u, p, w)
		}
	}
}

func TestHChain(t *testing.T) {
	hc := newHChain(2, 256)
	hc.put(0, pointer(0))
	hc.put(0, pointer(1))
	hc.resize(512)
	ptrs := make([]ptr, 4)
	n := hc.get(0, ptrs)
	if n != 2 {
		t.Fatalf("hc.put(0, p) returned %d; want %d", n, 2)
	}
	ptrs = ptrs[0:n]
	t.Log(ptrs)
	for _, p := range ptrs {
		if !(0 <= p.index() && p.index() <= 1) {
			t.Fatalf("p.index() is %d; want 0 or 1", p.index())
		}
	}
}

const example = `LZMA decoder test example
=========================
! LZMA ! Decoder ! TEST !
=========================
! TEST ! LZMA ! Decoder !
=========================
---- Test Line 1 --------
=========================
---- Test Line 2 --------
=========================
=== End of test file ====
=========================
`

func TestHCFinder(t *testing.T) {
	const depth = 20
	f, err := newHCFinder(4, 4096, 4096, 200, depth)
	if err != nil {
		t.Fatalf("newHCFinder error %s", err)
	}
	dict := f.Dict()
	n, err := io.WriteString(dict, example)
	if err != nil {
		t.Fatalf("io.WriteString(dict, example) error %s", err)
	}
	if n != len(example) {
		t.Fatalf("io.WriteString(dict, example) returned %d; want %d",
			n, len(example))
	}
	matches := make([]match, depth)
	t.Log("# Example")
	t.Log(example)
	for dict.Buffered() > 0 {
		pos := dict.Pos()
		end := pos + 10
		if end > int64(len(example)) {
			end = int64(len(example))
		}
		t.Logf("pos %d %q...\n", pos, example[pos:end])
		n := f.FindMatches(matches)
		if n == 0 {
			t.Log("no matches")
		}
		for i, m := range matches[:n] {
			t.Logf("matches[%d] = %v -> %q", i, m,
				example[pos:pos+int64(m.n)])
		}
		if n == 0 {
			f.Skip(1)
			continue
		}
		best := matches[n-1]
		f.Skip(best.n)
	}
	for i, ht := range f.ht {
		t.Logf("len(f.ht[%d].table) %d", i, len(ht.table))
	}
	t.Logf("len(f.hc.table) %d", len(f.hc.table))
}
