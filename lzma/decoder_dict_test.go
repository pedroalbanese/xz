package lzma

import (
	"io"
	"math/rand"
	"testing"
)

func fillRandom(d []byte, r *rand.Rand) {
	for i := range d {
		d[i] = byte(r.Int31n(256))
	}
}

func TestDecoderDict(t *testing.T) {
	r := rand.New(rand.NewSource(15))
	buf := make([]byte, 30)
	d, err := newDecoderDict(20, 10)
	if err != nil {
		t.Fatal("Couldn't create decoder dictionary.")
	}
	if cap(d.data) < 20 {
		t.Fatalf("cap(d.data) = %d; want at least %d", cap(d.data), 20)
	}
	t.Logf("d.data: [0:%d:%d]", len(d.data), cap(d.data))
	t.Logf("d %#v", d)
	buf = buf[:12]
	fillRandom(buf, r)
	n, err := d.Write(buf)
	if err != nil {
		t.Fatalf("d.Write(buf): %s", err)
	}
	if n != len(buf) {
		t.Fatalf("d.Write(buf) returned %d; want %d", n, len(buf))
	}
	if len(d.data) != n {
		t.Fatalf("len(d.data) = %d; want %d", len(d.data), n)
	}
	if d.c != n {
		t.Fatalf("d.c = %d; want %d", d.c, n)
	}
	if d.r != 0 {
		t.Fatalf("d.r = %d; want %d", d.r, 0)
	}
	buf = buf[:2]
	if n, err = d.Read(buf); err != nil {
		t.Fatalf("d.Read(buf): %s", err)
	}
	if n != 2 {
		t.Fatalf("d.Read(buf) = %d; want %d", n, 2)
	}
	t.Logf("d %#v", d)
	buf = buf[:19]
	fillRandom(buf, r)
	if n, err = d.Write(buf); err != nil {
		t.Fatalf("d.Write(buf) #2: %s", err)
	}
	if n != len(buf) {
		t.Fatalf("d.Write(buf) #2 = %d; want %d", n, len(buf))
	}
	t.Logf("d %#v", d)
	buf = buf[:19]
	if n, err = d.Read(buf); err != nil {
		t.Fatalf("d.Read(buf) #2: %s", err)
	}
	if n != 19 {
		t.Fatalf("d.Read(buf) #2 = %d; want %d", n, 19)
	}
}

func TestCopyMatch(t *testing.T) {
	r := rand.New(rand.NewSource(15))
	buf := make([]byte, 30)
	p, err := newDecoderDict(10, 10)
	if err != nil {
		t.Fatalf("newDecoderDict: %s", err)
	}
	t.Logf("cap(p.data): %d", cap(p.data))
	buf = buf[:5]
	fillRandom(buf, r)
	n, err := p.Write(buf)
	if err != nil {
		t.Fatalf("p.Write: %s\n", err)
	}
	if n != len(buf) {
		t.Fatalf("p.Write returned %d; want %d", n, len(buf))
	}
	t.Logf("p %#v", p)
	t.Log("copyMatch(2, 3)")
	if err = p.copyMatch(2, 3); err != nil {
		t.Fatal(err)
	}
	t.Logf("p %#v", p)
	t.Log("copyMatch(8, 8)")
	if err = p.copyMatch(8, 8); err != nil {
		t.Fatal(err)
	}
	t.Logf("p %#v", p)
	buf = buf[:30]
	if n, err = p.Read(buf); err != nil {
		t.Fatalf("Read: %s", err)
	}
	t.Logf("Read: %d", n)
	t.Log("copyMatch(2, 5)")
	if err = p.copyMatch(2, 5); err != nil {
		t.Fatal(err)
	}
	t.Logf("p %#v", p)
	if n, err = p.Read(buf); err != nil {
		t.Fatalf("Read: %s", err)
	}
	t.Logf("Read: %d", n)
	t.Log("copyMatch(2, 2)")
	if err = p.copyMatch(2, 2); err != nil {
		t.Fatal(err)
	}
	t.Logf("p %#v", p)
	if p.total != 23 {
		t.Fatalf("p.total %d; want %d", p.total, 23)
	}
}

func TestReset(t *testing.T) {
	p, err := newDecoderDict(10, 10)
	if err != nil {
		t.Fatalf("newDecoderDict: %s", err)
	}
	t.Logf("cap(p.data): %d", cap(p.data))
	r := rand.New(rand.NewSource(15))
	buf := make([]byte, 5)
	fillRandom(buf, r)
	n, err := p.Write(buf)
	if err != nil {
		t.Fatalf("p.Write: %s\n", err)
	}
	if n != len(buf) {
		t.Fatalf("p.Write returned %d; want %d", n, len(buf))
	}
	if p.total != 5 {
		t.Fatalf("p.total %d; want %d", p.total, 5)
	}
	p.reset()
	if p.total != 0 {
		t.Fatalf("p.total after reset %d; want %d", p.total, 0)
	}
	n = p.readable()
	if n != 0 {
		t.Fatalf("p.readable() after reset %d; want %d", n, 0)
	}
}

func TestDecoderDictEOF(t *testing.T) {
	p, err := newDecoderDict(10, 10)
	if err != nil {
		t.Fatalf("newDecoderDict: %s", err)
	}
	r := rand.New(rand.NewSource(15))
	buf := make([]byte, 5)
	fillRandom(buf, r)
	n, err := p.Write(buf)
	if err != nil {
		t.Fatalf("p.Write: %s\n", err)
	}
	if n != len(buf) {
		t.Fatalf("p.Write: returned %d; want %d", n, len(buf))
	}
	p.eof = true
	n, err = p.Read(buf)
	if err != nil {
		t.Fatalf("p.Read: error %s not expected", err)
	}
	if n != len(buf) {
		t.Fatalf("p.Read: got %d bytes; want %d", n, len(buf))
	}
	n, err = p.Read(buf)
	if err != io.EOF {
		t.Fatalf("p.Read: got err %s; want %s", err, io.EOF)
	}
	if n != 0 {
		t.Fatalf("p.Read: returned %d bytes; want %d", n, 0)
	}
}
