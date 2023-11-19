// SPDX-License-Identifier: Unlicense OR MIT

package ops

import (
	"encoding/binary"
)

// Reader parses an ops list.
type Reader struct {
	pc        PC
	stack     []macro
	ops       *Ops
	deferOps  Ops
	deferDone bool
}

// EncodedOp represents an encoded op returned by
// Reader.
type EncodedOp struct {
	Key  Key
	Data []byte
	Refs []interface{}
}

// Key is a unique key for a given op.
type Key struct {
	ops     *Ops
	pc      int
	version int
}

// Shadow of op.MacroOp.
type macroOp struct {
	ops   *Ops
	start PC
	end   PC
}

// PC is an instruction counter for an operation list.
type PC struct {
	data int
	refs int
}

type macro struct {
	ops   *Ops
	retPC PC
	endPC PC
}

type opMacroDef struct {
	endpc PC
}

func (pc PC) Add(op OpType) PC {
	size, numRefs := op.props()
	return PC{
		data: pc.data + size,
		refs: pc.refs + numRefs,
	}
}

// Reset start reading from the beginning of ops.
func (r *Reader) Reset(ops *Ops) {
	r.ResetAt(ops, PC{})
}

// ResetAt is like Reset, except it starts reading from pc.
func (r *Reader) ResetAt(ops *Ops, pc PC) {
	r.stack = r.stack[:0]
	Reset(&r.deferOps)
	r.deferDone = false
	r.pc = pc
	r.ops = ops
}

func (r *Reader) Decode() (EncodedOp, bool) {
	if r.ops == nil {
		return EncodedOp{}, false
	}
	deferring := false
	for {
		if len(r.stack) > 0 {
			b := r.stack[len(r.stack)-1]
			if r.pc == b.endPC {
				r.ops = b.ops
				r.pc = b.retPC
				r.stack = r.stack[:len(r.stack)-1]
				continue
			}
		}
		data := r.ops.data
		data = data[r.pc.data:]
		refs := r.ops.refs
		if len(data) == 0 {
			if r.deferDone {
				return EncodedOp{}, false
			}
			r.deferDone = true
			// Execute deferred macros.
			r.ops = &r.deferOps
			r.pc = PC{}
			continue
		}
		key := Key{ops: r.ops, pc: r.pc.data, version: r.ops.version}
		t := OpType(data[0])
		n, nrefs := t.props()
		data = data[:n]
		refs = refs[r.pc.refs:]
		refs = refs[:nrefs]
		switch t {
		case TypeDefer:
			deferring = true
			r.pc.data += n
			r.pc.refs += nrefs
			continue
		case TypeAux:
			// An Aux operations is always wrapped in a macro, and
			// its length is the remaining space.
			block := r.stack[len(r.stack)-1]
			n += block.endPC.data - r.pc.data - TypeAuxLen
			data = data[:n]
		case TypeCall:
			if deferring {
				deferring = false
				// Copy macro for deferred execution.
				if nrefs != 1 {
					panic("internal error: unexpected number of macro refs")
				}
				deferData := Write1(&r.deferOps, n, refs[0])
				copy(deferData, data)
				r.pc.data += n
				r.pc.refs += nrefs
				continue
			}
			var op macroOp
			op.decode(data, refs)
			retPC := r.pc
			retPC.data += n
			retPC.refs += nrefs
			r.stack = append(r.stack, macro{
				ops:   r.ops,
				retPC: retPC,
				endPC: op.end,
			})
			r.ops = op.ops
			r.pc = op.start
			continue
		case TypeMacro:
			var op opMacroDef
			op.decode(data)
			if op.endpc != (PC{}) {
				r.pc = op.endpc
			} else {
				// Treat an incomplete macro as containing all remaining ops.
				r.pc.data = len(r.ops.data)
				r.pc.refs = len(r.ops.refs)
			}
			continue
		}
		r.pc.data += n
		r.pc.refs += nrefs
		return EncodedOp{Key: key, Data: data, Refs: refs}, true
	}
}

func (op *opMacroDef) decode(data []byte) {
	if len(data) < TypeMacroLen || OpType(data[0]) != TypeMacro {
		panic("invalid op")
	}
	bo := binary.LittleEndian
	data = data[:TypeMacroLen]
	op.endpc.data = int(int32(bo.Uint32(data[1:])))
	op.endpc.refs = int(int32(bo.Uint32(data[5:])))
}

func (m *macroOp) decode(data []byte, refs []interface{}) {
	if len(data) < TypeCallLen || len(refs) < 1 || OpType(data[0]) != TypeCall {
		panic("invalid op")
	}
	bo := binary.LittleEndian
	data = data[:TypeCallLen]

	m.ops = refs[0].(*Ops)
	m.start.data = int(int32(bo.Uint32(data[1:])))
	m.start.refs = int(int32(bo.Uint32(data[5:])))
	m.end.data = int(int32(bo.Uint32(data[9:])))
	m.end.refs = int(int32(bo.Uint32(data[13:])))
}
