package akpool

import (
	"sync"
	"time"
)

const defaultLimit = 5000
const minInterval = time.Second / 3

func DefaultLimit() int { return defaultLimit }

type Item struct {
	AK       string `json:"ak"`
	Used     int    `json:"used"`
	Limit    int    `json:"limit"`
	Failed   bool   `json:"failed"`
	FailMsg  string `json:"failMsg"`
	lastUsed time.Time
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
		if i < len(limits) && limits[i] > 0 {
			item.Limit = limits[i]
		}
		p.items = append(p.items, item)
	}
	return p
}

func (p *Pool) Rebuild(aks []string, limits []int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	items := make([]*Item, len(aks))
	for i, ak := range aks {
		items[i] = &Item{AK: ak, Limit: defaultLimit}
		if i < len(limits) && limits[i] > 0 {
			items[i].Limit = limits[i]
		}
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
	p.mu.Lock()
	for _, item := range p.items {
		if item.AK == ak {
			elapsed := time.Since(item.lastUsed)
			if elapsed < minInterval {
				p.mu.Unlock()
				time.Sleep(minInterval - elapsed)
				return
			}
			item.lastUsed = time.Now()
			p.mu.Unlock()
			return
		}
	}
	p.mu.Unlock()
}

func (p *Pool) MarkFailed(ak, msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, item := range p.items {
		if item.AK == ak {
			item.Failed = true
			item.FailMsg = msg
			item.Used = item.Limit
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

func NeedsRotate(status int) bool {
	return status == 3 || status == 4 || status == 5 ||
		status == 200 || status == 201 || status == 202 ||
		status == 301 || status == 302
}
