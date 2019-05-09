package ninep

import "testing"

func TestPoolLimit(t *testing.T) {
	p := newPool(1000)
	for i := 0; i < 999; i++ {
		if _, ok := p.Get(); !ok {
			t.Fatalf("pool: limit reached after %d values", i)
		}
	}

	if _, ok := p.Get(); ok {
		t.Fatalf("pool: not exhausted when it should be")
	}
}

func TestPoolUnique(t *testing.T) {
	got := make(map[uint32]bool)
	p := newPool(5)

	for {
		v, ok := p.Get()
		if !ok {
			break
		}

		if _, ok := got[v]; ok {
			t.Fatalf("pool: found duplicate value %d", v)
		}
		got[v] = true
	}
}

func TestPoolRecycle(t *testing.T) {
	p := newPool(16)
	v1, _ := p.Get()
	p.Put(v1)
	v2, _ := p.Get()
	if v1 != v2 {
		t.Fatalf("pool: pool not recycled values")
	}
}
