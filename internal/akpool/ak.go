package akpool

import (
	"sync"
	"time"
)

const defaultLimit = 5000
const tokenRate = 2.0 // tokens per second
const tokenBurst = 2.0 // max tokens

func DefaultLimit() int { return defaultLimit }

type Item struct {
	Name     string `json:"name"`
	AK       string `json:"ak"`
	Used     int    `json:"used"`
	Limit    int    `json:"limit"`
	Failed   bool   `json:"failed"`
	FailMsg  string `json:"failMsg"`
	tokens   float64
	lastFill time.Time
}

type Pool struct {
	mu     sync.Mutex
	items  []*Item
	cursor int
}

func New(aks []string, limits []int) *Pool {
	p := &Pool{}
	for i, ak := range aks {
		item := &Item{AK: ak, Limit: defaultLimit}
		if i < len(limits) && limits[i] > 0 { item.Limit = limits[i] }
		p.items = append(p.items, item)
	}
	return p
}

func NewWithNames(aks []string, names []string, limits []int) *Pool {
	p := &Pool{}
	for i, ak := range aks {
		item := &Item{AK: ak, Name: ak, Limit: defaultLimit}
		if i < len(names) && names[i] != "" { item.Name = names[i] }
		if i < len(limits) && limits[i] > 0 { item.Limit = limits[i] }
		p.items = append(p.items, item)
	}
	return p
}

func (p *Pool) RebuildWithNames(aks []string, names []string, limits []int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	items := make([]*Item, len(aks))
	for i, ak := range aks {
		item := &Item{AK: ak, Name: ak, Limit: defaultLimit}
		if i < len(names) && names[i] != "" { item.Name = names[i] }
		if i < len(limits) && limits[i] > 0 { item.Limit = limits[i] }
		items[i] = item
	}
	p.items = items
	p.cursor = 0
}

func (p *Pool) Next() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.items) == 0 { return "" }
	for i := 0; i < len(p.items); i++ {
		idx := (p.cursor + i) % len(p.items)
		item := p.items[idx]
		if !item.Failed && item.Used < item.Limit {
			p.cursor = (idx + 1) % len(p.items)
			item.Used++
			return item.AK
		}
	}
	p.cursor = (p.cursor + 1) % len(p.items)
	return p.items[p.cursor].AK
}

func (p *Pool) Throttle(ak string) {
	for {
		p.mu.Lock()
		var item *Item
		for _, it := range p.items {
			if it.AK == ak {
				item = it
				break
			}
		}
		if item == nil {
			p.mu.Unlock()
			return
		}
		now := time.Now()
		if !item.lastFill.IsZero() {
			item.tokens = minFloat(tokenBurst, item.tokens+now.Sub(item.lastFill).Seconds()*tokenRate)
		} else {
			item.tokens = tokenBurst
		}
		item.lastFill = now
		if item.tokens >= 1 {
			item.tokens--
			p.mu.Unlock()
			return
		}
		need := time.Duration((1-item.tokens)/tokenRate*float64(time.Second))
		p.mu.Unlock()
		time.Sleep(need)
	}
}

func minFloat(a, b float64) float64 { if a < b { return a }; return b }

func (p *Pool) MarkFailed(ak, msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, item := range p.items {
		if item.AK == ak {
			item.Failed = true
			item.FailMsg = msg
			return
		}
	}
}

func (p *Pool) MarkSuccess(ak string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, item := range p.items {
		if item.AK == ak {
			item.Failed = false
			item.FailMsg = ""
			return
		}
	}
}

func (p *Pool) ResetAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, item := range p.items {
		item.Failed = false
		item.FailMsg = ""
		item.Used = 0
	}
}

func (p *Pool) Items() []Item {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Item, len(p.items))
	for i, item := range p.items {
		out[i] = *item
	}
	return out
}

func (p *Pool) AliveCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, item := range p.items {
		if !item.Failed && item.Used < item.Limit { n++ }
	}
	return n
}

// WorkerCount returns suggested parallel workers for one task (one per alive AK).
func (p *Pool) WorkerCount() int {
	n := p.AliveCount()
	if n < 1 {
		if len(p.items) > 0 {
			return 1
		}
		return 0
	}
	return n
}

func (p *Pool) Name(ak string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, item := range p.items {
		if item.AK == ak { return item.Name }
	}
	return ak[:minInt(8, len(ak))]
}

func NeedsRotate(status int) bool {
	return status == 3 || status == 4 || status == 5 ||
		status == 200 || status == 201 || status == 202 ||
		status == 301 || status == 302
}

func minInt(a, b int) int { if a < b { return a }; return b }
