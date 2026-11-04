package promtext

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// DefBuckets are the default histogram buckets, identical to client_golang's
// prometheus.DefBuckets.
var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// HistogramVec is a set of histograms sharing name, help and buckets,
// partitioned by label values. It is safe for concurrent use.
type HistogramVec struct {
	name       string
	help       string
	buckets    []float64
	labelNames []string

	mu       sync.Mutex
	children map[string]*histogramChild
}

type histogramChild struct {
	labels []Label
	counts []uint64 // per bucket, non-cumulative
	sum    float64
	count  uint64
}

// NewHistogramVec creates a histogram vector. Buckets must be strictly
// increasing; a trailing +Inf bucket is ignored since it is always implied.
func NewHistogramVec(name, help string, buckets []float64, labelNames ...string) *HistogramVec {
	b := append([]float64(nil), buckets...)
	if n := len(b); n > 0 && math.IsInf(b[n-1], +1) {
		b = b[:n-1]
	}
	for i := 1; i < len(b); i++ {
		if b[i] <= b[i-1] {
			panic(fmt.Sprintf("promtext: histogram %s buckets not strictly increasing", name))
		}
	}
	return &HistogramVec{
		name:       name,
		help:       help,
		buckets:    b,
		labelNames: append([]string(nil), labelNames...),
		children:   make(map[string]*histogramChild),
	}
}

// Observer records observations for one label combination.
type Observer struct {
	vec   *HistogramVec
	child *histogramChild
}

// WithLabelValues returns the observer for the given label values, which
// must match the label names in number and order.
func (v *HistogramVec) WithLabelValues(values ...string) Observer {
	if len(values) != len(v.labelNames) {
		panic(fmt.Sprintf("promtext: histogram %s expects %d label values, got %d", v.name, len(v.labelNames), len(values)))
	}
	key := strings.Join(values, "\xff")

	v.mu.Lock()
	defer v.mu.Unlock()
	c, ok := v.children[key]
	if !ok {
		labels := make([]Label, len(values))
		for i, val := range values {
			labels[i] = Label{Name: v.labelNames[i], Value: val}
		}
		c = &histogramChild{labels: labels, counts: make([]uint64, len(v.buckets))}
		v.children[key] = c
	}
	return Observer{vec: v, child: c}
}

// Observe adds a single observation. A value equal to a bucket's upper
// bound falls into that bucket.
func (o Observer) Observe(value float64) {
	i := sort.SearchFloat64s(o.vec.buckets, value)
	o.vec.mu.Lock()
	defer o.vec.mu.Unlock()
	if i < len(o.child.counts) {
		o.child.counts[i]++
	}
	o.child.sum += value
	o.child.count++
}

// Gather implements Gatherer. A vector without observations yields no family.
func (v *HistogramVec) Gather() []Family {
	v.mu.Lock()
	defer v.mu.Unlock()
	if len(v.children) == 0 {
		return nil
	}
	f := Family{Name: v.name, Help: v.help, Type: Histogram, Metrics: make([]Metric, 0, len(v.children))}
	for _, c := range v.children {
		h := &HistogramValue{Buckets: make([]Bucket, len(v.buckets)), Sum: c.sum, Count: c.count}
		var cum uint64
		for i, ub := range v.buckets {
			cum += c.counts[i]
			h.Buckets[i] = Bucket{UpperBound: ub, CumulativeCount: cum}
		}
		f.Metrics = append(f.Metrics, Metric{Labels: append([]Label(nil), c.labels...), Histogram: h})
	}
	return []Family{f}
}
