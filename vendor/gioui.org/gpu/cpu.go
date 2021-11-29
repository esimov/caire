// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"unsafe"

	"gioui.org/cpu"
)

// This file contains code specific to running compute shaders on the CPU.

// dispatcher dispatches CPU compute programs across multiple goroutines.
type dispatcher struct {
	// done is notified when a worker completes its work slice.
	done chan struct{}
	// work receives work slice indices. It is closed when the dispatcher is released.
	work chan work
	// dispatch receives compute jobs, which is then split among workers.
	dispatch chan dispatch
	// sync receives notification when a Sync completes.
	sync chan struct{}
}

type work struct {
	ctx   *cpu.DispatchContext
	index int
}

type dispatch struct {
	_type   jobType
	program *cpu.ProgramInfo
	descSet unsafe.Pointer
	x, y, z int
}

type jobType uint8

const (
	jobDispatch jobType = iota
	jobBarrier
	jobSync
)

func newDispatcher(workers int) *dispatcher {
	d := &dispatcher{
		work: make(chan work, workers),
		done: make(chan struct{}, workers),
		// Leave some room to avoid blocking calls to Dispatch.
		dispatch: make(chan dispatch, 20),
		sync:     make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		go d.worker()
	}
	go d.dispatcher()
	return d
}

func (d *dispatcher) dispatcher() {
	defer close(d.work)
	var free []*cpu.DispatchContext
	defer func() {
		for _, ctx := range free {
			ctx.Free()
		}
	}()
	var used []*cpu.DispatchContext
	for job := range d.dispatch {
		switch job._type {
		case jobDispatch:
			if len(free) == 0 {
				free = append(free, cpu.NewDispatchContext())
			}
			ctx := free[len(free)-1]
			free = free[:len(free)-1]
			used = append(used, ctx)
			ctx.Prepare(cap(d.work), job.program, job.descSet, job.x, job.y, job.z)
			for i := 0; i < cap(d.work); i++ {
				d.work <- work{
					ctx:   ctx,
					index: i,
				}
			}
		case jobBarrier:
			// Wait for all outstanding dispatches to complete.
			for i := 0; i < len(used)*cap(d.work); i++ {
				<-d.done
			}
			free = append(free, used...)
			used = used[:0]
		case jobSync:
			d.sync <- struct{}{}
		}
	}
}

func (d *dispatcher) worker() {
	thread := cpu.NewThreadContext()
	defer thread.Free()
	for w := range d.work {
		w.ctx.Dispatch(w.index, thread)
		d.done <- struct{}{}
	}
}

func (d *dispatcher) Barrier() {
	d.dispatch <- dispatch{_type: jobBarrier}
}

func (d *dispatcher) Sync() {
	d.dispatch <- dispatch{_type: jobSync}
	<-d.sync
}

func (d *dispatcher) Dispatch(program *cpu.ProgramInfo, descSet unsafe.Pointer, x, y, z int) {
	d.dispatch <- dispatch{
		_type:   jobDispatch,
		program: program,
		descSet: descSet,
		x:       x,
		y:       y,
		z:       z,
	}
}

func (d *dispatcher) Stop() {
	close(d.dispatch)
}
