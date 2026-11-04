package promtext

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, g Gatherer) string {
	t.Helper()
	var sb strings.Builder
	require.NoError(t, Write(&sb, g.Gather()))
	return sb.String()
}

func TestWriteGaugesSortedAndEscaped(t *testing.T) {
	fams := []Family{
		{Name: "b_metric", Help: "line1\nback\\slash", Type: Gauge, Metrics: []Metric{
			{Labels: []Label{{"z", "1"}, {"a", "q\"\\\n"}}, Value: 2},
			{Labels: []Label{{"z", "0"}, {"a", "q\"\\\n"}}, Value: math.Inf(1)},
		}},
		{Name: "a_metric", Help: "h", Type: Gauge, Metrics: []Metric{{Value: 0.5}}},
		{Name: "empty", Help: "dropped", Type: Gauge},
	}
	var sb strings.Builder
	require.NoError(t, Write(&sb, fams))
	assert.Equal(t, `# HELP a_metric h
# TYPE a_metric gauge
a_metric 0.5
# HELP b_metric line1\nback\\slash
# TYPE b_metric gauge
b_metric{a="q\"\\\n",z="0"} +Inf
b_metric{a="q\"\\\n",z="1"} 2
`, sb.String())
}

func TestGatherersMergeFamilies(t *testing.T) {
	g := Gatherers{
		GathererFunc(func() []Family {
			return []Family{{Name: "m", Help: "first", Type: Gauge, Metrics: []Metric{{Labels: []Label{{"k", "b"}}, Value: 1}}}}
		}),
		nil,
		GathererFunc(func() []Family {
			return []Family{{Name: "m", Help: "second", Type: Gauge, Metrics: []Metric{{Labels: []Label{{"k", "a"}}, Value: 2}}}}
		}),
	}
	assert.Equal(t, "# HELP m first\n# TYPE m gauge\nm{k=\"a\"} 2\nm{k=\"b\"} 1\n", render(t, g))
}

func TestHistogramBucketsAreCumulativeAndInclusive(t *testing.T) {
	v := NewHistogramVec("h", "help", []float64{1, 2, math.Inf(1)}, "l")
	o := v.WithLabelValues("x")
	for _, val := range []float64{0, 1, 1.5, 2, 3} {
		o.Observe(val)
	}
	assert.Equal(t, `# HELP h help
# TYPE h histogram
h_bucket{l="x",le="1"} 2
h_bucket{l="x",le="2"} 4
h_bucket{l="x",le="+Inf"} 5
h_sum{l="x"} 7.5
h_count{l="x"} 5
`, render(t, v))
}

func TestHistogramWithoutObservationsIsOmitted(t *testing.T) {
	assert.Empty(t, render(t, NewHistogramVec("h", "help", DefBuckets)))
}

func TestHistogramPanicsOnBadInput(t *testing.T) {
	assert.Panics(t, func() { NewHistogramVec("h", "", []float64{2, 1}) })
	assert.Panics(t, func() { NewHistogramVec("h", "", DefBuckets, "a").WithLabelValues() })
}

func TestHistogramConcurrentObserve(t *testing.T) {
	v := NewHistogramVec("h", "help", []float64{1}, "worker")
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				v.WithLabelValues([]string{"a", "b"}[w%2]).Observe(0.5)
				_ = v.Gather()
			}
		}(w)
	}
	wg.Wait()
	out := render(t, v)
	assert.Contains(t, out, `h_count{worker="a"} 4000`)
	assert.Contains(t, out, `h_count{worker="b"} 4000`)
}

func TestHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(GathererFunc(func() []Family {
		return []Family{{Name: "g", Help: "h", Type: Gauge, Metrics: []Metric{{Value: 1}}}}
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Equal(t, ContentType, rec.Header().Get("Content-Type"))
	assert.Equal(t, "# HELP g h\n# TYPE g gauge\ng 1\n", rec.Body.String())
}
