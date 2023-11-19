package harfbuzz

// Code generated with ragel -Z -o ot_khmer_machine.go ot_khmer_machine.rl ; sed -i '/^\/\/line/ d' ot_khmer_machine.go ; goimports -w ot_khmer_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-khmer-machine.rl Copyright © 2015 Google, Inc. Behdad Esfahbod

const (
	khmerConsonantSyllable = iota
	khmerBrokenCluster
	khmerNonKhmerCluster
)

const khmSM_ex_C = 1
const khmSM_ex_DOTTEDCIRCLE = 11
const khmSM_ex_H = 4
const khmSM_ex_PLACEHOLDER = 10
const khmSM_ex_Ra = 15
const khmSM_ex_Robatic = 25
const khmSM_ex_V = 2
const khmSM_ex_VAbv = 20
const khmSM_ex_VBlw = 21
const khmSM_ex_VPre = 22
const khmSM_ex_VPst = 23
const khmSM_ex_Xgroup = 26
const khmSM_ex_Ygroup = 27
const khmSM_ex_ZWJ = 6
const khmSM_ex_ZWNJ = 5

var _khmSM_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 5,
	1, 6, 1, 7, 1, 8, 1, 9,
	1, 10, 1, 11, 2, 2, 3, 2,
	2, 4,
}

var _khmSM_key_offsets []byte = []byte{
	0, 5, 8, 11, 15, 18, 21, 25,
	28, 32, 35, 40, 45, 48, 51, 55,
	58, 61, 65, 68, 72, 75, 90, 100,
	103, 113, 122, 123, 129, 134, 141, 149,
	158, 168, 171, 181, 190, 191, 197, 202,
	209, 217, 226,
}

var _khmSM_trans_keys []byte = []byte{
	20, 25, 26, 5, 6, 26, 5, 6,
	15, 1, 2, 20, 26, 5, 6, 26,
	5, 6, 26, 5, 6, 20, 26, 5,
	6, 26, 5, 6, 20, 26, 5, 6,
	26, 5, 6, 20, 25, 26, 5, 6,
	20, 25, 26, 5, 6, 26, 5, 6,
	15, 1, 2, 20, 26, 5, 6, 26,
	5, 6, 26, 5, 6, 20, 26, 5,
	6, 26, 5, 6, 20, 26, 5, 6,
	26, 5, 6, 4, 15, 20, 21, 22,
	23, 25, 26, 27, 1, 2, 5, 6,
	10, 11, 4, 20, 21, 22, 23, 25,
	26, 27, 5, 6, 15, 1, 2, 4,
	20, 21, 22, 23, 25, 26, 27, 5,
	6, 4, 20, 21, 22, 23, 26, 27,
	5, 6, 27, 4, 23, 26, 27, 5,
	6, 4, 26, 27, 5, 6, 4, 20,
	23, 26, 27, 5, 6, 4, 20, 21,
	23, 26, 27, 5, 6, 4, 20, 21,
	22, 23, 26, 27, 5, 6, 4, 20,
	21, 22, 23, 25, 26, 27, 5, 6,
	15, 1, 2, 4, 20, 21, 22, 23,
	25, 26, 27, 5, 6, 4, 20, 21,
	22, 23, 26, 27, 5, 6, 27, 4,
	23, 26, 27, 5, 6, 4, 26, 27,
	5, 6, 4, 20, 23, 26, 27, 5,
	6, 4, 20, 21, 23, 26, 27, 5,
	6, 4, 20, 21, 22, 23, 26, 27,
	5, 6, 20, 26, 5, 6,
}

var _khmSM_single_lengths []byte = []byte{
	3, 1, 1, 2, 1, 1, 2, 1,
	2, 1, 3, 3, 1, 1, 2, 1,
	1, 2, 1, 2, 1, 9, 8, 1,
	8, 7, 1, 4, 3, 5, 6, 7,
	8, 1, 8, 7, 1, 4, 3, 5,
	6, 7, 2,
}

var _khmSM_range_lengths []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 3, 1, 1,
	1, 1, 0, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 0, 1, 1, 1,
	1, 1, 1,
}

var _khmSM_index_offsets []int16 = []int16{
	0, 5, 8, 11, 15, 18, 21, 25,
	28, 32, 35, 40, 45, 48, 51, 55,
	58, 61, 65, 68, 72, 75, 88, 98,
	101, 111, 120, 122, 128, 133, 140, 148,
	157, 167, 170, 180, 189, 191, 197, 202,
	209, 217, 226,
}

var _khmSM_indicies []byte = []byte{
	2, 3, 4, 1, 0, 4, 1, 0,
	5, 5, 0, 2, 4, 1, 0, 2,
	6, 0, 8, 7, 0, 2, 10, 9,
	0, 10, 9, 0, 2, 12, 11, 0,
	12, 11, 0, 2, 13, 4, 1, 0,
	16, 17, 18, 15, 14, 18, 15, 19,
	20, 20, 14, 16, 18, 15, 14, 16,
	21, 14, 23, 22, 14, 16, 25, 24,
	14, 25, 24, 14, 16, 27, 26, 14,
	27, 26, 14, 30, 29, 16, 25, 27,
	23, 17, 18, 20, 29, 31, 13, 28,
	33, 2, 10, 12, 8, 13, 4, 5,
	34, 32, 35, 35, 32, 33, 2, 10,
	12, 8, 3, 4, 5, 36, 32, 37,
	2, 10, 12, 8, 4, 5, 38, 32,
	5, 32, 37, 8, 2, 5, 6, 32,
	37, 8, 5, 7, 32, 37, 2, 8,
	10, 5, 39, 32, 37, 2, 10, 8,
	12, 5, 40, 32, 33, 2, 10, 12,
	8, 4, 5, 38, 32, 33, 2, 10,
	12, 8, 3, 4, 5, 38, 32, 42,
	42, 41, 30, 16, 25, 27, 23, 17,
	18, 20, 43, 41, 44, 16, 25, 27,
	23, 18, 20, 45, 41, 20, 41, 44,
	23, 16, 20, 21, 41, 44, 23, 20,
	22, 41, 44, 16, 23, 25, 20, 46,
	41, 44, 16, 25, 23, 27, 20, 47,
	41, 30, 16, 25, 27, 23, 18, 20,
	45, 41, 16, 18, 15, 48,
}

var _khmSM_trans_targs []byte = []byte{
	21, 1, 27, 31, 25, 26, 4, 5,
	28, 7, 29, 9, 30, 32, 21, 12,
	37, 41, 35, 21, 36, 15, 16, 38,
	18, 39, 20, 40, 21, 22, 33, 42,
	21, 23, 10, 24, 0, 2, 3, 6,
	8, 21, 34, 11, 13, 14, 17, 19,
	21,
}

var _khmSM_trans_actions []byte = []byte{
	15, 0, 5, 5, 5, 0, 0, 0,
	5, 0, 5, 0, 5, 5, 17, 0,
	5, 21, 21, 19, 0, 0, 0, 5,
	0, 5, 0, 5, 7, 5, 0, 24,
	9, 0, 0, 5, 0, 0, 0, 0,
	0, 11, 21, 0, 0, 0, 0, 0,
	13,
}

var _khmSM_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0,
}

var _khmSM_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 3, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0,
}

var _khmSM_eof_trans []int16 = []int16{
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 15, 20, 15, 15, 15,
	15, 15, 15, 15, 15, 0, 33, 33,
	33, 33, 33, 33, 33, 33, 33, 33,
	33, 42, 42, 42, 42, 42, 42, 42,
	42, 42, 49,
}

const khmSM_start int = 21
const khmSM_first_final int = 21
const khmSM_error int = -1

const khmSM_en_main int = 21

func findSyllablesKhmer(buffer *Buffer) {
	var p, ts, te, act, cs int
	info := buffer.Info

	{
		cs = khmSM_start
		ts = 0
		te = 0
		act = 0
	}

	pe := len(info)
	eof := pe

	var syllableSerial uint8 = 1

	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if p == pe {
			goto _test_eof
		}
	_resume:
		_acts = int(_khmSM_from_state_actions[cs])
		_nacts = uint(_khmSM_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _khmSM_actions[_acts-1] {
			case 1:
				ts = p

			}
		}

		_keys = int(_khmSM_key_offsets[cs])
		_trans = int(_khmSM_index_offsets[cs])

		_klen = int(_khmSM_single_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + _klen - 1)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + ((_upper - _lower) >> 1)
				switch {
				case (info[p].complexCategory) < _khmSM_trans_keys[_mid]:
					_upper = _mid - 1
				case (info[p].complexCategory) > _khmSM_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_khmSM_range_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case (info[p].complexCategory) < _khmSM_trans_keys[_mid]:
					_upper = _mid - 2
				case (info[p].complexCategory) > _khmSM_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_khmSM_indicies[_trans])
	_eof_trans:
		cs = int(_khmSM_trans_targs[_trans])

		if _khmSM_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_khmSM_trans_actions[_trans])
		_nacts = uint(_khmSM_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _khmSM_actions[_acts-1] {
			case 2:
				te = p + 1

			case 3:
				act = 2
			case 4:
				act = 3
			case 5:
				te = p + 1
				{
					foundSyllableKhmer(khmerNonKhmerCluster, ts, te, info, &syllableSerial)
				}
			case 6:
				te = p
				p--
				{
					foundSyllableKhmer(khmerConsonantSyllable, ts, te, info, &syllableSerial)
				}
			case 7:
				te = p
				p--
				{
					foundSyllableKhmer(khmerBrokenCluster, ts, te, info, &syllableSerial)
					buffer.scratchFlags |= bsfHasBrokenSyllable
				}
			case 8:
				te = p
				p--
				{
					foundSyllableKhmer(khmerNonKhmerCluster, ts, te, info, &syllableSerial)
				}
			case 9:
				p = (te) - 1
				{
					foundSyllableKhmer(khmerConsonantSyllable, ts, te, info, &syllableSerial)
				}
			case 10:
				p = (te) - 1
				{
					foundSyllableKhmer(khmerBrokenCluster, ts, te, info, &syllableSerial)
					buffer.scratchFlags |= bsfHasBrokenSyllable
				}
			case 11:
				switch act {
				case 2:
					{
						p = (te) - 1
						foundSyllableKhmer(khmerBrokenCluster, ts, te, info, &syllableSerial)
						buffer.scratchFlags |= bsfHasBrokenSyllable
					}
				case 3:
					{
						p = (te) - 1
						foundSyllableKhmer(khmerNonKhmerCluster, ts, te, info, &syllableSerial)
					}
				}

			}
		}

	_again:
		_acts = int(_khmSM_to_state_actions[cs])
		_nacts = uint(_khmSM_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _khmSM_actions[_acts-1] {
			case 0:
				ts = 0

			}
		}

		p++
		if p != pe {
			goto _resume
		}
	_test_eof:
		{
		}
		if p == eof {
			if _khmSM_eof_trans[cs] > 0 {
				_trans = int(_khmSM_eof_trans[cs] - 1)
				goto _eof_trans
			}
		}

	}

	_ = act // needed by Ragel, but unused
}
