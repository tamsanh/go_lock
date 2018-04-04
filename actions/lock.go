package actions

import (
	"github.com/gobuffalo/buffalo"
	"sync"
	"time"
	"fmt"
	"strconv"
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
	seconds := params.Get("stale_after")
	secondsInt := int64(0)
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	secondsWasSet := false
	if seconds != "" {
		secondsInt, _ = strconv.ParseInt(seconds, 10, 64)
		secondsWasSet = true
	}

	namedLocks.mux.Lock()
	defer namedLocks.mux.Unlock()

	renderString := "failed"

	// name gets locked, seconds is set
	nameIsLocked := namedLocks.isLocked[name]
	if secondsWasSet && nameIsLocked {
		fmt.Printf("Failed Lock %s: Still locked\n", name)
		renderString = "failed"
	} else if secondsWasSet && !nameIsLocked {
		// Check if we are within timeout time
		lastBeat := namedLocks.beats[name]
		timeDifference := time.Now().Sub(lastBeat)
		if int64(timeDifference.Seconds()) > secondsInt {
			staleness := secondsInt - int64(timeDifference.Seconds())
			namedLocks.isLocked[name] = true
			namedLocks.beats[name] = time.Now()
			fmt.Printf("Locked %s: Stale by %s Seconds\n", name, staleness)
			renderString = "success"
		} else {
			freshness := int64(timeDifference.Seconds()) - secondsInt
			fmt.Printf("Ignore Lock %s: Fresh by %s seconds\n", name, freshness)
			renderString = "fresh"
		}
	} else if !secondsWasSet && !nameIsLocked {
		namedLocks.isLocked[name] = true
		namedLocks.beats[name] = time.Now()
		fmt.Printf("Locked %s\n", name)
		renderString = "success"
	} else if !secondsWasSet && nameIsLocked {
		fmt.Printf("Failed Lock %s: Still locked\n", name)
		renderString = "failed"
	}

	return c.Render(200, r.String(renderString))
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

