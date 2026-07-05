package core

import (
	"strconv"
	"testing"
)

func TestCacheRegionPutGet(t *testing.T) {
	c := NewCache()
	r := c.Region(RegionDNS)
	r.Put("example.com", "93.184.216.34")
	v, ok := r.Get("example.com")
	if !ok || v.(string) != "93.184.216.34" {
		t.Fatalf("get: ok=%v v=%v", ok, v)
	}
}

func TestCacheLRUEviction(t *testing.T) {
	c := NewCache()
	// Use a fresh custom region via Region() then overflow a small one.
	r := newRegion(3)
	for i := 0; i < 5; i++ {
		r.Put(strconv.Itoa(i), i)
	}
	if r.Len() != 3 {
		t.Fatalf("expected len 3 after eviction, got %d", r.Len())
	}
	// Oldest (0,1) should be gone; (2,3,4) present.
	if _, ok := r.Get("0"); ok {
		t.Fatal("key 0 should have been evicted")
	}
	if _, ok := r.Get("4"); !ok {
		t.Fatal("key 4 should be present")
	}
	_ = c
}
