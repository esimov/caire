package uax

import (
	"fmt"

	"github.com/gioui/uax/internal/tracing"
)

// DefaultRunePublisher is a type to organize RuneSubscribers.
//
// Rune publishers have to maintain a list of subscribers. Subscribers are
// then notified on the arrival of new runes (code-points) by sending
// them rune-events. When a subscriber is done with consuming runes (subscribers
// are often short-lived), it signals Done()=true.
//
// The DefaultRunePublisher data structure "prioritizes" subscribers with
// Done()=true within a queue.
// It maintains a "gap" position between done and not-done. The queue grows as
// needed.
//
// A DefaultRunePublisher implements RunePublisher.
type DefaultRunePublisher struct {
	q   []RuneSubscriber // queue is slice of subscribers
	gap int              // index of first subscriber which is Done(), may be out of range
	//aggregate      PenaltyAggregator // see declaration of RunePublisher
	penaltiesTotal []int // set of penalties collected from subscribers
}

// Len returns the number of subscribers held.
func (pq DefaultRunePublisher) Len() int { return len(pq.q) }

func (pq DefaultRunePublisher) empty() bool { return len(pq.q) == 0 }

func (pq DefaultRunePublisher) at(i int) RuneSubscriber {
	return pq.q[i]
}

// Top subscriber in queue. If there is at last one Done() subscriber, top()
// will return one.
func (pq DefaultRunePublisher) Top() RuneSubscriber {
	if pq.Len() == 0 {
		return nil
	}
	return pq.q[pq.Len()-1]
}

// Fix signals that the
// Done()-flag of a subscriber has changed: inform the queue to let
// it re-organize.
func (pq *DefaultRunePublisher) Fix(at int) {
	if at < pq.Len() {
		//pq.print()
		if pq.q[at].Done() {
			pq.bubbleUp(at)
		} else {
			pq.bubbleDown(at)
		}
		for i := 0; i < pq.gap; i++ {
			if pq.q[i].Done() {
				tracing.Errorf("prioq.Fix(%d) failed", at)
				pq.print()
				panic("internal queue order compromised")
			}
		}
	}
}

// Push puts a new subscriber to the queue.
func (pq *DefaultRunePublisher) Push(subscr RuneSubscriber) {
	l := pq.Len() // index of new item
	pq.q = append(pq.q, subscr)
	if !pq.Top().Done() {
		pq.bubbleDown(l)
	}
	//fmt.Printf("#### length of prio queue = %d\n", pq.Len())
}

// Pop the topmost subscriber.
func (pq *DefaultRunePublisher) Pop() RuneSubscriber {
	if pq == nil || pq.Len() == 0 {
		return nil
	}
	old := pq.q
	n := len(old)
	subscr := old[n-1]
	pq.q = old[0 : n-1]
	old[n-1] = nil
	if pq.gap > pq.Len() {
		pq.gap--
	}
	return subscr
}

// PopDone pops the topmost subscriber if it is Done(), otherwise return nil.
// If the method returns nil, the queue either is empty or holds
// subscribers with Done()=false only (i.e., subscribers still active).
func (pq *DefaultRunePublisher) PopDone() RuneSubscriber {
	if pq == nil || pq.Len() == 0 {
		return nil
	}
	if pq.Top().Done() {
		return pq.Pop()
	}
	return nil
}

// Pre-requisite: subscriber at positition is Done().
func (pq *DefaultRunePublisher) bubbleUp(i int) {
	if i < pq.gap-1 {
		if pq.gap < pq.Len() {
			pq.q[i], pq.q[pq.gap-1] = pq.q[pq.gap-1], pq.q[i] // swap
		} else if i < pq.Len()-1 { // gap is out of range
			last := pq.Len() - 1
			pq.q[i], pq.q[last] = pq.q[last], pq.q[i] // swap with topmost
		}
	}
	if pq.gap > 0 && pq.q[pq.gap-1].Done() {
		pq.gap--
	}
}

// Pre-requisite: subscriber at positition is not Done().
func (pq *DefaultRunePublisher) bubbleDown(i int) {
	if i >= pq.gap {
		if pq.gap < i {
			pq.q[i], pq.q[pq.gap] = pq.q[pq.gap], pq.q[i] // swap
		}
	}
	if pq.gap < pq.Len() && !pq.q[pq.gap].Done() {
		pq.gap++
	}
}

func (pq *DefaultRunePublisher) print() {
	fmt.Printf("Publisher of length %d (gap = %d):\n", pq.Len(), pq.gap)
	for i := pq.Len() - 1; i >= 0; i-- {
		subscr := pq.at(i)
		fmt.Printf(" - [%d] %s done=%v\n", i, subscr, subscr.Done())
	}
}
