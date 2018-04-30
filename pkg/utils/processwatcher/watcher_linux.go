package processwatcher

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func PathToInode(p string) (uint64, error) {

	var stat syscall.Stat_t
	if err := syscall.Stat(p, &stat); err != nil {
		return 0, err
	}
	return stat.Ino, nil
}

var (
	attached = make(map[int]bool)
	rwlock   sync.RWMutex
)

type Watcher struct {
	inod uint64
	pids map[int]bool
}

func NewWatcher(inod uint64) *Watcher {
	return &Watcher{
		inod: inod,
		pids: make(map[int]bool),
	}
}

func (w *Watcher) Watch() <-chan int {
	c := make(chan int)
	go func() {
		for {
			pids, err := FindPids(w.inod)
			if err != nil {
				panic("error getting pids")
			}

			// stale old pids
		outloop:
			for pid := range w.pids {
				for _, newpid := range pids {
					if pid == newpid {
						continue outloop
					}
				}
				delete(w.pids, pid)
			}
			// notify on new pids
			for _, pid := range pids {
				if _, ok := w.pids[pid]; !ok {
					log.WithFields(log.Fields{"pid": pid}).Debug("match found")
					w.pids[pid] = true
					c <- pid
				}
			}
			time.Sleep(time.Second)
		}
	}()
	return c
}

func FindPids(inod uint64) ([]int, error) {
	var res []int
	files, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}

		//		log.WithFields(log.Fields{"inod": inod, "pid": pid}).Debug("Testing")

		p := filepath.Join("/proc", f.Name(), "exe")
		if inod2, err := PathToInode(p); err != nil {
			//		log.WithFields(log.Fields{"err": err, "pid": pid}).Debug("match Error found")
			continue
		} else if inod == inod2 {
			//	log.WithFields(log.Fields{"pid": pid}).Debug("match found")
			res = append(res, pid)
		} else {
			//		log.WithFields(log.Fields{"inod2": inod2, "pid": pid}).Debug("match NOT found")
		}
	}

	return res, nil
}
