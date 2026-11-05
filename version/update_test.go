package version

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

const httpAgent = "pkg/test"

func TestNewUpdate(t *testing.T) {
	version = "1.0.0"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "10.0.0")
	}))
	defer svr.Close()
	versionURL = svr.URL

	wg := &sync.WaitGroup{}
	wg.Add(1)

	onUpdate := false
	u := NewUpdateAndStart(httpAgent)
	defer u.StopWatch()
	u.SetOnUpdateListener(func() {
		onUpdate = true
		wg.Done()
	})

	waitTimeout(wg)
	if onUpdate != true {
		t.Errorf("update not found")
	}
}

func TestDoNotUpdate(t *testing.T) {
	version = "11.0.0"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "10.0.0")
	}))
	defer svr.Close()
	versionURL = svr.URL

	wg := &sync.WaitGroup{}
	wg.Add(1)

	onUpdate := false
	u := NewUpdateAndStart(httpAgent)
	defer u.StopWatch()
	u.SetOnUpdateListener(func() {
		onUpdate = true
		wg.Done()
	})

	waitTimeout(wg)
	if onUpdate == true {
		t.Errorf("invalid update")
	}
}

func TestDaemonUpdate(t *testing.T) {
	version = "11.0.0"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "11.0.0")
	}))
	defer svr.Close()
	versionURL = svr.URL

	wg := &sync.WaitGroup{}
	wg.Add(1)

	onUpdate := false
	u := NewUpdateAndStart(httpAgent)
	defer u.StopWatch()
	u.SetOnUpdateListener(func() {
		onUpdate = true
		wg.Done()
	})

	u.SetDaemonVersion("10.0.0")

	waitTimeout(wg)
	if onUpdate != true {
		t.Errorf("invalid daemon version check")
	}
}

func waitTimeout(wg *sync.WaitGroup) {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	select {
	case <-c:
		return
	case <-time.After(time.Second):
		return
	}
}

func TestDoNotUpdateForkPrerelease(t *testing.T) {
	version = "0.82.0-ktn.2"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "0.82.0")
	}))
	defer svr.Close()
	versionURL = svr.URL

	wg := &sync.WaitGroup{}
	wg.Add(1)

	onUpdate := false
	u := NewUpdateAndStart(httpAgent)
	defer u.StopWatch()
	u.SetOnUpdateListener(func() {
		onUpdate = true
		wg.Done()
	})

	waitTimeout(wg)
	if onUpdate {
		t.Errorf("fork prerelease should not offer update for same numeric upstream version")
	}
}

func TestUpdateWhenUpstreamNumericHigher(t *testing.T) {
	version = "0.82.0-ktn.2"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "0.83.0")
	}))
	defer svr.Close()
	versionURL = svr.URL

	wg := &sync.WaitGroup{}
	wg.Add(1)

	onUpdate := false
	u := NewUpdateAndStart(httpAgent)
	defer u.StopWatch()
	u.SetOnUpdateListener(func() {
		onUpdate = true
		wg.Done()
	})

	waitTimeout(wg)
	if !onUpdate {
		t.Errorf("expected update when upstream numeric version is higher")
	}
}
