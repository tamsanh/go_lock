package actions

import (
	"github.com/gobuffalo/buffalo"
	"sync"
	"time"
	"fmt"
)


type NamedLocks struct {
	isLocked map[string]bool
	beats map[string]time.Time
	mux sync.Mutex
}


var namedLocks = NamedLocks{isLocked:make(map[string]bool), beats: make(map[string]time.Time)}


func (nl NamedLocks) unlockName(name string) {
	fmt.Println("Unlocking name " + name)
	nl.mux.Lock()
	nl.isLocked[name] = false
	nl.mux.Unlock()
}


func (nl NamedLocks) beat(name string) {
	nl.mux.Lock()
	fmt.Println("Beating " + name)
	nl.beats[name] = time.Now()
	nl.mux.Unlock()
}


func (nl NamedLocks) isStale(name string, timeout_seconds int) bool {
	timeDifference := time.Now().Sub(nl.beats[name])
	return int(timeDifference.Seconds()) >= timeout_seconds
}


func UnlockStaleLocks(timeout_seconds int) {
	checkInterval := time.Second * 30
	for {
		time.Sleep(checkInterval)
		namedLocks.mux.Lock()
		for name, isLocked := range namedLocks.isLocked {
			if isLocked {
				if namedLocks.isStale(name, timeout_seconds) {
					fmt.Println("Unlocked " + name)
					namedLocks.isLocked[name] = false
					namedLocks.beats[name] = time.Now()
				}
			}
		}
		namedLocks.mux.Unlock()
	}
}


// LockCreate default implementation.
func LockCreate(c buffalo.Context) error {
	params := c.Params()
	name := params.Get("name")
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	namedLocks.mux.Lock()
	defer namedLocks.mux.Unlock()

	if namedLocks.isLocked[name] {
		return c.Render(200, r.String("failed"))
	} else {
		fmt.Println("Locking name " + name)
		namedLocks.isLocked[name] = true
		namedLocks.beats[name] = time.Now()
		return c.Render(200, r.String("success"))
	}
}

// LockHeartbeat default implementation.
func LockHeartbeat(c buffalo.Context) error {
	params := c.Params()
	name := params.Get("name")
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	namedLocks.beat(name)
	return c.Render(200, r.String(name))
}

// LockStatus default implementation.
func LockStatus(c buffalo.Context) error {
	params := c.Params()
	name := params.Get("name")
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	namedLocks.mux.Lock()
	defer namedLocks.mux.Unlock()
	isLocked, exists := namedLocks.isLocked[name]
	if isLocked {
		return c.Render(200, r.String("locked"))
	} else {
		if exists {
			lastTime := namedLocks.beats[name]
			return c.Render(200, r.String(lastTime.UTC().Format(time.RFC3339)))
		} else {
			return c.Render(200, r.String("failed"))
		}
	}
}
