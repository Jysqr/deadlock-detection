package Barrier

import (
	"errors"
	"fmt"
	"sync"
)

// the motivation for this is to be able to have all the nodes fire at once, then wait for a GUI signal from the GUI node. This will allow the UI to be updated in a sane way
//gist of it is, n nodes each get a mutex that link back to this bad boy, main node gets the trigger to kick it forward
type barrier struct {
	syncList      []*sync.Mutex
	Ready         bool
	currentReturn int
}

func NewBarrier(total int) barrier {
	nb := barrier{
		syncList:      make([]*sync.Mutex, total),
		Ready:         true,
		currentReturn: 0,
	}
	for i := 0; i < total; i++ {
		nb.syncList[i] = &sync.Mutex{}
	}
	return nb
}

func (b *barrier) Step() error {
	var err error
	if b.Ready {
		b.Ready = false
		fmt.Println("unlocking")
		wg := sync.WaitGroup{}
		for _, value := range b.syncList {
			wg.Add(1)
			go func(value *sync.Mutex) { //this spits into a goroutine so each mutex doesnt lock the thread while it waits to reclaim
				value.Unlock()
				value.Lock()
				wg.Done()
			}(value)
		}
		fmt.Println("locked")
		wg.Wait()
		b.Ready = true
		err = nil
	} else {
		err = errors.New("all nodes have not completed their step")
	}
	return err
}

func (b *barrier) Mutex() *sync.Mutex {
	m := b.syncList[b.currentReturn]
	m.Lock() //prelocks the mutex so the goroutine will wait on start
	b.currentReturn++
	return m
}
