// Package transfer contains operations and events for brokering data transfers.
//
// The transfer protocol is as follows:
//
//  - Data sources are registered with SourceOps, data targets with TargetOps.
//  - A data source receives a RequestEvent when a transfer is initiated.
//    It must respond with an OfferOp.
//  - The target receives a DataEvent when transferring to it. It must close
//    the event data after use.
//
// When a user initiates a pointer-guided drag and drop transfer, the
// source as well as all potential targets receive an InitiateEvent.
// Potential targets are targets with at least one MIME type in common
// with the source. When a drag gesture completes, a CancelEvent is sent
// to the source and all potential targets.
//
// Note that the RequestEvent is sent to the source upon drop.
package transfer

import (
	"io"

	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/op"
)

// SourceOp registers a tag as a data source for a MIME type.
// Use multiple SourceOps if a tag supports multiple types.
type SourceOp struct {
	Tag event.Tag
	// Type is the MIME type supported by this source.
	Type string
}

// TargetOp registers a tag as a data target.
// Use multiple TargetOps if a tag supports multiple types.
type TargetOp struct {
	Tag event.Tag
	// Type is the MIME type accepted by this target.
	Type string
}

// OfferOp is used by data sources as a response to a RequestEvent.
type OfferOp struct {
	Tag event.Tag
	// Type is the MIME type of Data.
	// It must be the Type from the corresponding RequestEvent.
	Type string
	// Data contains the offered data. It is closed when the
	// transfer is complete or cancelled.
	// Data must be kept valid until closed, and it may be used from
	// a goroutine separate from the one processing the frame..
	Data io.ReadCloser
}

func (op SourceOp) Add(o *op.Ops) {
	data := ops.Write2(&o.Internal, ops.TypeSourceLen, op.Tag, op.Type)
	data[0] = byte(ops.TypeSource)
}

func (op TargetOp) Add(o *op.Ops) {
	data := ops.Write2(&o.Internal, ops.TypeTargetLen, op.Tag, op.Type)
	data[0] = byte(ops.TypeTarget)
}

// Add the offer to the list of operations.
// It panics if the Data field is not set.
func (op OfferOp) Add(o *op.Ops) {
	if op.Data == nil {
		panic("invalid nil data in OfferOp")
	}
	data := ops.Write3(&o.Internal, ops.TypeOfferLen, op.Tag, op.Type, op.Data)
	data[0] = byte(ops.TypeOffer)
}

// RequestEvent requests data from a data source. The source must
// respond with an OfferOp.
type RequestEvent struct {
	// Type is the first matched type between the source and the target.
	Type string
}

func (RequestEvent) ImplementsEvent() {}

// InitiateEvent is sent to a data source when a drag-and-drop
// transfer gesture is initiated.
//
// Potential data targets also receive the event.
type InitiateEvent struct{}

func (InitiateEvent) ImplementsEvent() {}

// CancelEvent is sent to data sources and targets to cancel the
// effects of an InitiateEvent.
type CancelEvent struct{}

func (CancelEvent) ImplementsEvent() {}

// DataEvent is sent to the target receiving the transfer.
type DataEvent struct {
	// Type is the MIME type of Data.
	Type string
	// Open returns the transfer data. It is only valid to call Open in the frame
	// the DataEvent is received. The caller must close the return value after use.
	Open func() io.ReadCloser
}

func (DataEvent) ImplementsEvent() {}
