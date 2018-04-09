package actions

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
)

type NamedLocks struct {
	isLocked map[string]bool
	beats    map[string]time.Time
	mux      sync.Mutex
}

var namedLocks = NamedLocks{isLocked: make(map[string]bool), beats: make(map[string]time.Time)}

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

func (nl NamedLocks) isStale(name string, timeoutSeconds int) bool {
	timeDifference := time.Now().Sub(nl.beats[name])
	return int(timeDifference.Seconds()) >= timeoutSeconds
}

func DestroyStaleLocks(timeoutSeconds int) {
	checkInterval := time.Second * 30
	for {
		time.Sleep(checkInterval)
		namedLocks.mux.Lock()
		for name, isLocked := range namedLocks.isLocked {
			if isLocked {
				if namedLocks.isStale(name, timeoutSeconds) {
					fmt.Println("Destroyed stale " + name)
					delete(namedLocks.isLocked, name)
					delete(namedLocks.beats, name)
				}
			}
		}
		namedLocks.mux.Unlock()
	}
}

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
			staleness := int64(timeDifference.Seconds()) - secondsInt
			namedLocks.isLocked[name] = true
			namedLocks.beats[name] = time.Now()
			fmt.Printf("Locked %s: Stale by %d Seconds\n", name, staleness)
			renderString = "success"
		} else {
			freshness := secondsInt - int64(timeDifference.Seconds())
			fmt.Printf("Ignore Lock %s: Fresh by %d seconds\n", name, freshness)
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

func LockHeartbeat(c buffalo.Context) error {
	params := c.Params()
	name := params.Get("name")
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	if namedLocks.isLocked[name] {
		namedLocks.beat(name)
	}
	return c.Render(200, r.String(name))
}

func LockDestroy(c buffalo.Context) error {
	params := c.Params()
	name := params.Get("name")
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	namedLocks.mux.Lock()
	defer namedLocks.mux.Unlock()
	delete(namedLocks.beats, name)
	delete(namedLocks.isLocked, name)
	fmt.Println("Destroyed " + name)
	return c.Render(200, r.String("success"))
}

func LockUnlock(c buffalo.Context) error {
	params := c.Params()
	name := params.Get("name")
	if name == "" {
		return c.Render(200, r.String("failed"))
	}
	namedLocks.mux.Lock()
	defer namedLocks.mux.Unlock()
	namedLocks.beats[name] = time.Now()
	namedLocks.isLocked[name] = false
	fmt.Println("Unlocked " + name)
	return c.Render(200, r.String("success"))
}
