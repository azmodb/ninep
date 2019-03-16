package ninep

import "testing"

func TestPool(t *testing.T) {
	p := newPool(1, 10)
	var cache []uint32
	for i := 1; i < 10; i++ {
		v, ok := p.Next()
		if !ok {
			t.Fatalf("pool: exhausted before 10 cache values (%d)", i)
		}
		cache = append(cache, v)
	}

	if _, ok := p.Next(); ok {
		t.Fatalf("pool: not exhausted when it should be")
	}

	if len(p.cache) != 0 {
		t.Fatalf("pool: expected empyt cache, have %d", len(p.cache))
	}

	for _, v := range cache {
		p.Put(v)
	}

	if len(cache) != len(p.cache) {
		t.Fatalf("pool: unepxecpted cache size %d", len(p.cache))
	}
}

func TestPoolRecycle(t *testing.T) {
	p := newPool(1, 3)
	v1, _ := p.Next()
	p.Put(v1)
	v2, _ := p.Next()
	if v1 != v2 {
		t.Errorf("pool not recycling values")
	}
}

func TestPoolUnique(t *testing.T) {
	got := make(map[uint32]struct{})
	p := newPool(1, 500)

	for {
		v, ok := p.Next()
		if !ok {
			break
		}

		if _, ok := got[v]; ok {
			t.Errorf("pool: found duplicate value %d", v)
		}
		got[v] = struct{}{}
	}
}
