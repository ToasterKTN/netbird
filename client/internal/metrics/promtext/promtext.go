// Package promtext is a minimal, dependency-free implementation of the
// Prometheus text exposition format (version 0.0.4) covering the metric types
// the client exposes on its local /metrics endpoint: gauges and histograms.
package promtext

import (
	"bufio"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// ContentType is the Content-Type of the text exposition format. It matches
// what client_golang's promhttp handler sends for text responses.
const ContentType = "text/plain; version=0.0.4; charset=utf-8; escaping=underscores"

// Type is a metric family type.
type Type string

const (
	// Gauge is a single value that can go up and down.
	Gauge Type = "gauge"
	// Histogram is a set of cumulative buckets with a sum and count.
	Histogram Type = "histogram"
)

// Label is a single name/value label pair.
type Label struct {
	Name  string
	Value string
}

// Bucket is one cumulative histogram bucket.
type Bucket struct {
	UpperBound      float64
	CumulativeCount uint64
}

// HistogramValue is a histogram snapshot. Buckets must be sorted by upper
// bound and must not include +Inf; the +Inf bucket is derived from Count.
type HistogramValue struct {
	Buckets []Bucket
	Sum     float64
	Count   uint64
}

// Metric is one labeled sample. Value is used for gauges and Histogram for
// histograms.
type Metric struct {
	Labels    []Label
	Value     float64
	Histogram *HistogramValue
}

// Family is a named group of metrics sharing help text and type.
type Family struct {
	Name    string
	Help    string
	Type    Type
	Metrics []Metric
}

// Gatherer produces metric families on demand.
type Gatherer interface {
	Gather() []Family
}

// GathererFunc adapts a function to Gatherer.
type GathererFunc func() []Family

// Gather implements Gatherer.
func (f GathererFunc) Gather() []Family { return f() }

// Gatherers combines several gatherers. Families with the same name are
// merged; the first help text and type win.
type Gatherers []Gatherer

// Gather implements Gatherer.
func (gs Gatherers) Gather() []Family {
	var out []Family
	for _, g := range gs {
		if g == nil {
			continue
		}
		out = append(out, g.Gather()...)
	}
	return out
}

// Handler serves the families of g in the text exposition format.
func Handler(g Gatherer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", ContentType)
		_ = Write(w, g.Gather())
	})
}

// Write encodes families in the text exposition format. Output is stable:
// families are merged by name and sorted, labels are sorted by name, and
// metrics are sorted by their label values. Families without metrics are
// omitted.
func Write(w io.Writer, families []Family) error {
	bw := bufio.NewWriter(w)
	for _, f := range normalize(families) {
		writeFamily(bw, f)
	}
	return bw.Flush()
}

func normalize(families []Family) []Family {
	byName := make(map[string]int, len(families))
	var merged []Family
	for _, f := range families {
		if i, ok := byName[f.Name]; ok {
			merged[i].Metrics = append(merged[i].Metrics, f.Metrics...)
			continue
		}
		byName[f.Name] = len(merged)
		f.Metrics = append([]Metric(nil), f.Metrics...)
		merged = append(merged, f)
	}

	out := merged[:0]
	for _, f := range merged {
		if len(f.Metrics) == 0 {
			continue
		}
		for i := range f.Metrics {
			labels := append([]Label(nil), f.Metrics[i].Labels...)
			sort.SliceStable(labels, func(a, b int) bool { return labels[a].Name < labels[b].Name })
			f.Metrics[i].Labels = labels
		}
		sort.SliceStable(f.Metrics, func(a, b int) bool { return lessLabels(f.Metrics[a].Labels, f.Metrics[b].Labels) })
		out = append(out, f)
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Name < out[b].Name })
	return out
}

func lessLabels(a, b []Label) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i].Name != b[i].Name {
			return a[i].Name < b[i].Name
		}
		if a[i].Value != b[i].Value {
			return a[i].Value < b[i].Value
		}
	}
	return len(a) < len(b)
}

func writeFamily(w *bufio.Writer, f Family) {
	w.WriteString("# HELP ")
	w.WriteString(f.Name)
	w.WriteByte(' ')
	w.WriteString(helpEscaper.Replace(f.Help))
	w.WriteString("\n# TYPE ")
	w.WriteString(f.Name)
	w.WriteByte(' ')
	w.WriteString(string(f.Type))
	w.WriteByte('\n')

	for _, m := range f.Metrics {
		if f.Type != Histogram || m.Histogram == nil {
			writeSample(w, f.Name, m.Labels, nil, m.Value)
			continue
		}
		h := m.Histogram
		for _, b := range h.Buckets {
			le := Label{Name: "le", Value: formatFloat(b.UpperBound)}
			writeSample(w, f.Name+"_bucket", m.Labels, &le, float64(b.CumulativeCount))
		}
		inf := Label{Name: "le", Value: "+Inf"}
		writeSample(w, f.Name+"_bucket", m.Labels, &inf, float64(h.Count))
		writeSample(w, f.Name+"_sum", m.Labels, nil, h.Sum)
		writeSample(w, f.Name+"_count", m.Labels, nil, float64(h.Count))
	}
}

func writeSample(w *bufio.Writer, name string, labels []Label, extra *Label, value float64) {
	w.WriteString(name)
	if len(labels) > 0 || extra != nil {
		w.WriteByte('{')
		sep := false
		write := func(l Label) {
			if sep {
				w.WriteByte(',')
			}
			sep = true
			w.WriteString(l.Name)
			w.WriteString(`="`)
			w.WriteString(labelEscaper.Replace(l.Value))
			w.WriteByte('"')
		}
		for _, l := range labels {
			write(l)
		}
		if extra != nil {
			write(*extra)
		}
		w.WriteByte('}')
	}
	w.WriteByte(' ')
	w.WriteString(formatFloat(value))
	w.WriteByte('\n')
}

var (
	labelEscaper = strings.NewReplacer(`\`, `\\`, "\n", `\n`, `"`, `\"`)
	helpEscaper  = strings.NewReplacer(`\`, `\\`, "\n", `\n`)
)

func formatFloat(f float64) string {
	switch {
	case math.IsInf(f, +1):
		return "+Inf"
	case math.IsInf(f, -1):
		return "-Inf"
	case math.IsNaN(f):
		return "NaN"
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}
