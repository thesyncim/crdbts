package crdbts

import (
	"sync"
)

var analyserPool sync.Pool

func AcquireAnalyser() *Analyser {
	if v := analyserPool.Get(); v != nil {
		return v.(*Analyser)
	}
	return NewAnalyser()
}

func ReleaseAnalyser(a *Analyser) {
	analyserPool.Put(a)
}
