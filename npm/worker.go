package npm

import (
	"net/url"
	"sync"

	"github.com/ssut/pocketnpm/db"
	"github.com/ssut/pocketnpm/log"
)

// MirrorWorker contains channels used to act as a worker
type MirrorWorker struct {
	ID          int
	Work        chan *db.BarePackage
	WorkerQueue chan chan *db.BarePackage
	ResultQueue chan *MirrorWorkResult
	WaitGroup   *sync.WaitGroup
	QuitChan    chan bool

	npmClient *NPMClient
}

// MirrorWorkResult contains the result of worker action
type MirrorWorkResult struct {
	Package  *db.BarePackage
	Document string
	Files    []*url.URL
	Ok       bool
}

// NewMirrorWorker creates a worker with given parameters
func NewMirrorWorker(id int, npmClient *NPMClient, workerQueue chan chan *db.BarePackage, resultQueue chan *MirrorWorkResult, wg *sync.WaitGroup) *MirrorWorker {
	worker := &MirrorWorker{
		ID:          id,
		Work:        make(chan *db.BarePackage),
		WorkerQueue: workerQueue,
		ResultQueue: resultQueue,
		WaitGroup:   wg,
		QuitChan:    make(chan bool),
		npmClient:   npmClient,
	}

	return worker
}

// Start method starts the worker by starting a goroutine
func (w *MirrorWorker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// log.Printf("Worker %d received work request: %s %s", w.ID, work.ID, work.Revision)
				w.ResultQueue <- &MirrorWorkResult{
					Package: work,
					Ok:      true,
				}
				w.WaitGroup.Done()
			case <-w.QuitChan:
				log.Printf("Worker %d Stopping", w.ID)
				w.WaitGroup.Done()
				return
			}
		}
	}()
}

// Stop function tells the worker to stop listening for work requests
//
// Note that the worker will only stop after it has finished its work
func (w *MirrorWorker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
