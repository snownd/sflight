package sflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	g := New[int, string](5 * time.Second)
	v, err, _ := g.Do(1, func() (string, error) {
		return "sflight", nil
	})
	if v != "sflight" {
		t.Errorf("v = %v, want %v", v, "sflight")
	}
	if err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
}

func TestDoErr(t *testing.T) {
	g := New[int, string](5 * time.Second)
	wantErr := errors.New("test error")
	_, err, _ := g.Do(1, func() (string, error) {
		return "sflight", wantErr
	})
	if err != wantErr {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func TestDoConcurrent(t *testing.T) {
	g := New[int, string](5 * time.Second)
	wg := sync.WaitGroup{}
	executed := atomic.Int32{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err, _ := g.Do(1, func() (string, error) {
				executed.Add(1)
				time.Sleep(10 * time.Millisecond)
				return "a", nil
			})
			if v != "a" {
				t.Errorf("v = %v, want %v", v, "a")
			}
			if err != nil {
				t.Errorf("err = %v, want %v", err, nil)
			}
		}()
	}
	wg.Wait()
	if executed.Load() != 1 {
		t.Errorf("executed = %v, want %v", executed.Load(), 1)
	}
}

func TestDoConcurrentExpired(t *testing.T) {
	g := New[int, string](100 * time.Millisecond)
	wg := sync.WaitGroup{}
	executed := atomic.Int32{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err, _ := g.Do(1, func() (string, error) {
				executed.Add(1)
				time.Sleep(10 * time.Millisecond)
				return "a", nil
			})
			if v != "a" {
				t.Errorf("v = %v, want %v", v, "a")
			}
			if err != nil {
				t.Errorf("err = %v, want %v", err, nil)
			}
		}()
	}
	wg.Wait()
	if executed.Load() != 1 {
		t.Errorf("executed = %v, want %v", executed.Load(), 1)
	}
	time.Sleep(200 * time.Millisecond)
	v, _, isExecuted := g.Do(1, func() (string, error) {
		return "b", nil
	})
	if v != "b" {
		t.Errorf("v = %v, want %v", v, "b")
	}
	if !isExecuted {
		t.Errorf("isExecuted = %v, want %v", isExecuted, true)
	}
}

func TestForget(t *testing.T) {
	g := New[int, string](5 * time.Second)
	g.Do(1, func() (string, error) {
		return "a", nil
	})
	g.Forget(1)
	v, _, isExecuted := g.Do(1, func() (string, error) {
		return "b", nil
	})
	if v != "b" {
		t.Errorf("v = %v, want %v", v, "b")
	}
	if !isExecuted {
		t.Errorf("isExecuted = %v, want %v", isExecuted, true)
	}
}
