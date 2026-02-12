package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Registry struct {
	mu        sync.Mutex
	counters  map[string]float64
	histogram map[string][]float64
	buckets   []float64
}

func New() *Registry {
	return &Registry{
		counters:  map[string]float64{},
		histogram: map[string][]float64{},
		buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 8},
	}
}

func (r *Registry) Inc(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[key(name, labels)]++
}

func (r *Registry) Observe(name string, labels map[string]string, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(name, labels)
	if _, ok := r.histogram[k]; !ok {
		r.histogram[k] = make([]float64, len(r.buckets)+1)
	}
	secs := d.Seconds()
	for i, b := range r.buckets {
		if secs <= b {
			r.histogram[k][i]++
		}
	}
	r.histogram[k][len(r.buckets)]++
}

func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		for _, line := range r.render() {
			_, _ = w.Write([]byte(line + "\n"))
		}
	})
}

func (r *Registry) render() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	lines := make([]string, 0)
	keys := make([]string, 0, len(r.counters))
	for k := range r.counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s %v", k, r.counters[k]))
	}
	hk := make([]string, 0, len(r.histogram))
	for k := range r.histogram {
		hk = append(hk, k)
	}
	sort.Strings(hk)
	for _, k := range hk {
		vals := r.histogram[k]
		for i, b := range r.buckets {
			lines = append(lines, fmt.Sprintf("%s_bucket{le=\"%.2f\"} %v", k, b, vals[i]))
		}
		lines = append(lines, fmt.Sprintf("%s_bucket{le=\"+Inf\"} %v", k, vals[len(vals)-1]))
	}
	return lines
}

func key(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	sort.Strings(parts)
	return fmt.Sprintf("%s{%s}", sanitize(name), strings.Join(parts, ","))
}

func sanitize(n string) string {
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.ReplaceAll(n, ".", "_")
	return n
}
