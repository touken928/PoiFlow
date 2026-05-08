package akpool

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	p := New([]string{"ak1", "ak2"}, nil)
	if p.AliveCount() != 2 {
		t.Errorf("expected 2 alive, got %d", p.AliveCount())
	}
}

func TestNextRotation(t *testing.T) {
	p := New([]string{"ak1", "ak2", "ak3"}, nil)
	seen := make(map[string]int)
	for i := 0; i < 6; i++ {
		ak := p.Next()
		seen[ak]++
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 unique AKs, got %d", len(seen))
	}
	for _, count := range seen {
		if count != 2 {
			t.Errorf("expected each AK used 2 times, got %d", count)
		}
	}
}

func TestMarkFailed(t *testing.T) {
	p := New([]string{"ak1", "ak2"}, nil)
	p.MarkFailed("ak1", "配额不足")
	items := p.Items()

	if !items[0].Failed {
		t.Error("expected ak1 to be marked failed")
	}
	if items[0].FailMsg != "配额不足" {
		t.Errorf("expected fail msg '配额不足', got '%s'", items[0].FailMsg)
	}
	if p.AliveCount() != 1 {
		t.Errorf("expected 1 alive, got %d", p.AliveCount())
	}
}

func TestAllFailed(t *testing.T) {
	p := New([]string{"ak1", "ak2"}, nil)
	p.MarkFailed("ak1", "err")
	p.MarkFailed("ak2", "err")
	// Even if all failed, Next should still return something (fallback)
	ak := p.Next()
	if ak == "" {
		t.Error("expected non-empty AK even if all failed")
	}
	if p.AliveCount() != 0 {
		t.Errorf("expected 0 alive, got %d", p.AliveCount())
	}
}

func TestResetAll(t *testing.T) {
	p := New([]string{"ak1", "ak2"}, nil)
	p.MarkFailed("ak1", "err")
	p.ResetAll()
	for _, item := range p.Items() {
		if item.Failed {
			t.Error("expected all items to be reset")
		}
	}
}

func TestNeedsRotate(t *testing.T) {
	tests := []struct {
		status int
		expect bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, true},
		{4, true},
		{5, true},
		{200, true},
		{201, true},
		{202, true},
		{301, true},
		{302, true},
	}

	for _, tt := range tests {
		t.Run(strings.Join([]string{"status_", itoa(tt.status)}, ""), func(t *testing.T) {
			if NeedsRotate(tt.status) != tt.expect {
				t.Errorf("NeedsRotate(%d) = %v, want %v", tt.status, !tt.expect, tt.expect)
			}
		})
	}
}

func TestItems(t *testing.T) {
	p := New([]string{"ak1", "ak2"}, []int{100, 200})
	items := p.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].AK != "ak1" || items[0].Limit != 100 {
		t.Errorf("unexpected item 0: %+v", items[0])
	}
	if items[1].AK != "ak2" || items[1].Limit != 200 {
		t.Errorf("unexpected item 1: %+v", items[1])
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
