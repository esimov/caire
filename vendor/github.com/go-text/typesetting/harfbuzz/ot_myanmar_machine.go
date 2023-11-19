package harfbuzz

// Code generated with ragel -Z -o ot_myanmar_machine.go ot_myanmar_machine.rl ; sed -i '/^\/\/line/ d' ot_myanmar_machine.go ; goimports -w ot_myanmar_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-myanmar-machine.rl Copyright © 2015 Mozilla Foundation. Google, Inc. Behdad Esfahbod

// myanmar_syllable_type_t
const (
	myanmarConsonantSyllable = iota
	myanmarBrokenCluster
	myanmarNonMyanmarCluster
)

const myaSM_ex_A = 9
const myaSM_ex_As = 32
const myaSM_ex_C = 1
const myaSM_ex_CS = 18
const myaSM_ex_DB = 3
const myaSM_ex_DOTTEDCIRCLE = 11
const myaSM_ex_GB = 10
const myaSM_ex_H = 4
const myaSM_ex_IV = 2
const myaSM_ex_MH = 35
const myaSM_ex_ML = 41
const myaSM_ex_MR = 36
const myaSM_ex_MW = 37
const myaSM_ex_MY = 38
const myaSM_ex_PT = 39
const myaSM_ex_Ra = 15
const myaSM_ex_SM = 8
const myaSM_ex_VAbv = 20
const myaSM_ex_VBlw = 21
const myaSM_ex_VPre = 22
const myaSM_ex_VPst = 23
const myaSM_ex_VS = 40
const myaSM_ex_ZWJ = 6
const myaSM_ex_ZWNJ = 5

var _myaSM_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3,
	1, 4, 1, 5, 1, 6, 1, 7,
	1, 8,
}

var _myaSM_key_offsets []int16 = []int16{
	0, 24, 42, 48, 51, 62, 69, 76,
	81, 85, 93, 102, 112, 117, 120, 129,
	137, 148, 158, 174, 186, 197, 210, 223,
	238, 252, 269, 275, 278, 289, 296, 303,
	308, 312, 320, 329, 339, 344, 347, 365,
	374, 382, 393, 403, 419, 431, 442, 455,
	468, 483, 497, 514, 532, 549, 572,
}

var _myaSM_trans_keys []byte = []byte{
	3, 4, 8, 9, 15, 18, 20, 21,
	22, 23, 32, 35, 36, 37, 38, 39,
	40, 41, 1, 2, 5, 6, 10, 11,
	3, 4, 8, 9, 20, 21, 22, 23,
	32, 35, 36, 37, 38, 39, 40, 41,
	5, 6, 8, 23, 32, 39, 5, 6,
	8, 5, 6, 3, 8, 9, 20, 23,
	32, 35, 39, 41, 5, 6, 3, 8,
	9, 23, 39, 5, 6, 3, 8, 9,
	32, 39, 5, 6, 8, 32, 39, 5,
	6, 8, 39, 5, 6, 3, 8, 9,
	20, 23, 39, 5, 6, 3, 8, 9,
	20, 23, 32, 39, 5, 6, 3, 8,
	9, 20, 23, 32, 39, 41, 5, 6,
	8, 23, 39, 5, 6, 15, 1, 2,
	3, 8, 9, 20, 21, 23, 39, 5,
	6, 3, 8, 9, 21, 23, 39, 5,
	6, 3, 8, 9, 20, 21, 22, 23,
	39, 40, 5, 6, 3, 8, 9, 20,
	21, 22, 23, 39, 5, 6, 3, 8,
	9, 20, 21, 22, 23, 32, 35, 36,
	37, 38, 39, 41, 5, 6, 3, 8,
	9, 20, 21, 22, 23, 32, 39, 41,
	5, 6, 3, 8, 9, 20, 21, 22,
	23, 32, 39, 5, 6, 3, 8, 9,
	20, 21, 22, 23, 35, 37, 39, 41,
	5, 6, 3, 8, 9, 20, 21, 22,
	23, 32, 35, 39, 41, 5, 6, 3,
	8, 9, 20, 21, 22, 23, 32, 35,
	36, 37, 39, 41, 5, 6, 3, 8,
	9, 20, 21, 22, 23, 35, 36, 37,
	39, 41, 5, 6, 3, 4, 8, 9,
	20, 21, 22, 23, 32, 35, 36, 37,
	38, 39, 41, 5, 6, 8, 23, 32,
	39, 5, 6, 8, 5, 6, 3, 8,
	9, 20, 23, 32, 35, 39, 41, 5,
	6, 3, 8, 9, 23, 39, 5, 6,
	3, 8, 9, 32, 39, 5, 6, 8,
	32, 39, 5, 6, 8, 39, 5, 6,
	3, 8, 9, 20, 23, 39, 5, 6,
	3, 8, 9, 20, 23, 32, 39, 5,
	6, 3, 8, 9, 20, 23, 32, 39,
	41, 5, 6, 8, 23, 39, 5, 6,
	15, 1, 2, 3, 4, 8, 9, 20,
	21, 22, 23, 32, 35, 36, 37, 38,
	39, 40, 41, 5, 6, 3, 8, 9,
	20, 21, 23, 39, 5, 6, 3, 8,
	9, 21, 23, 39, 5, 6, 3, 8,
	9, 20, 21, 22, 23, 39, 40, 5,
	6, 3, 8, 9, 20, 21, 22, 23,
	39, 5, 6, 3, 8, 9, 20, 21,
	22, 23, 32, 35, 36, 37, 38, 39,
	41, 5, 6, 3, 8, 9, 20, 21,
	22, 23, 32, 39, 41, 5, 6, 3,
	8, 9, 20, 21, 22, 23, 32, 39,
	5, 6, 3, 8, 9, 20, 21, 22,
	23, 35, 37, 39, 41, 5, 6, 3,
	8, 9, 20, 21, 22, 23, 32, 35,
	39, 41, 5, 6, 3, 8, 9, 20,
	21, 22, 23, 32, 35, 36, 37, 39,
	41, 5, 6, 3, 8, 9, 20, 21,
	22, 23, 35, 36, 37, 39, 41, 5,
	6, 3, 4, 8, 9, 20, 21, 22,
	23, 32, 35, 36, 37, 38, 39, 41,
	5, 6, 3, 4, 8, 9, 20, 21,
	22, 23, 32, 35, 36, 37, 38, 39,
	40, 41, 5, 6, 3, 4, 8, 9,
	20, 21, 22, 23, 32, 35, 36, 37,
	38, 39, 41, 5, 6, 3, 4, 8,
	9, 15, 20, 21, 22, 23, 32, 35,
	36, 37, 38, 39, 40, 41, 1, 2,
	5, 6, 10, 11, 15, 1, 2, 10,
	11,
}

var _myaSM_single_lengths []byte = []byte{
	18, 16, 4, 1, 9, 5, 5, 3,
	2, 6, 7, 8, 3, 1, 7, 6,
	9, 8, 14, 10, 9, 11, 11, 13,
	12, 15, 4, 1, 9, 5, 5, 3,
	2, 6, 7, 8, 3, 1, 16, 7,
	6, 9, 8, 14, 10, 9, 11, 11,
	13, 12, 15, 16, 15, 17, 1,
}

var _myaSM_range_lengths []byte = []byte{
	3, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 3, 2,
}

var _myaSM_index_offsets []int16 = []int16{
	0, 22, 40, 46, 49, 60, 67, 74,
	79, 83, 91, 100, 110, 115, 118, 127,
	135, 146, 156, 172, 184, 195, 208, 221,
	236, 250, 267, 273, 276, 287, 294, 301,
	306, 310, 318, 327, 337, 342, 345, 363,
	372, 380, 391, 401, 417, 429, 440, 453,
	466, 481, 495, 512, 530, 547, 568,
}

var _myaSM_indicies []byte = []byte{
	2, 3, 5, 6, 7, 8, 9, 10,
	11, 12, 13, 14, 15, 16, 17, 18,
	19, 20, 1, 4, 1, 0, 22, 23,
	25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 24, 21,
	25, 30, 39, 36, 24, 21, 25, 24,
	21, 22, 25, 26, 40, 30, 41, 42,
	36, 41, 24, 21, 22, 25, 26, 30,
	36, 24, 21, 43, 25, 36, 44, 36,
	24, 21, 25, 44, 36, 24, 21, 25,
	36, 24, 21, 22, 25, 26, 40, 30,
	36, 24, 21, 22, 25, 26, 40, 30,
	41, 36, 24, 21, 22, 25, 26, 40,
	30, 41, 36, 41, 24, 21, 25, 30,
	36, 24, 21, 1, 1, 21, 22, 25,
	26, 27, 28, 30, 36, 24, 21, 22,
	25, 26, 28, 30, 36, 24, 21, 22,
	25, 26, 27, 28, 29, 30, 36, 45,
	24, 21, 22, 25, 26, 27, 28, 29,
	30, 36, 24, 21, 22, 25, 26, 27,
	28, 29, 30, 31, 32, 33, 34, 35,
	36, 38, 24, 21, 22, 25, 26, 27,
	28, 29, 30, 45, 36, 38, 24, 21,
	22, 25, 26, 27, 28, 29, 30, 45,
	36, 24, 21, 22, 25, 26, 27, 28,
	29, 30, 32, 34, 36, 38, 24, 21,
	22, 25, 26, 27, 28, 29, 30, 45,
	32, 36, 38, 24, 21, 22, 25, 26,
	27, 28, 29, 30, 46, 32, 33, 34,
	36, 38, 24, 21, 22, 25, 26, 27,
	28, 29, 30, 32, 33, 34, 36, 38,
	24, 21, 22, 23, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36,
	38, 24, 21, 5, 12, 49, 18, 48,
	47, 5, 48, 47, 2, 5, 6, 50,
	12, 51, 52, 18, 51, 48, 47, 2,
	5, 6, 12, 18, 48, 47, 53, 5,
	18, 54, 18, 48, 47, 5, 54, 18,
	48, 47, 5, 18, 48, 47, 2, 5,
	6, 50, 12, 18, 48, 47, 2, 5,
	6, 50, 12, 51, 18, 48, 47, 2,
	5, 6, 50, 12, 51, 18, 51, 48,
	47, 5, 12, 18, 48, 47, 55, 55,
	47, 2, 3, 5, 6, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19,
	20, 48, 47, 2, 5, 6, 9, 10,
	12, 18, 48, 47, 2, 5, 6, 10,
	12, 18, 48, 47, 2, 5, 6, 9,
	10, 11, 12, 18, 56, 48, 47, 2,
	5, 6, 9, 10, 11, 12, 18, 48,
	47, 2, 5, 6, 9, 10, 11, 12,
	13, 14, 15, 16, 17, 18, 20, 48,
	47, 2, 5, 6, 9, 10, 11, 12,
	56, 18, 20, 48, 47, 2, 5, 6,
	9, 10, 11, 12, 56, 18, 48, 47,
	2, 5, 6, 9, 10, 11, 12, 14,
	16, 18, 20, 48, 47, 2, 5, 6,
	9, 10, 11, 12, 56, 14, 18, 20,
	48, 47, 2, 5, 6, 9, 10, 11,
	12, 57, 14, 15, 16, 18, 20, 48,
	47, 2, 5, 6, 9, 10, 11, 12,
	14, 15, 16, 18, 20, 48, 47, 2,
	3, 5, 6, 9, 10, 11, 12, 13,
	14, 15, 16, 17, 18, 20, 48, 47,
	22, 23, 25, 26, 27, 28, 29, 30,
	58, 32, 33, 34, 35, 36, 37, 38,
	24, 21, 22, 59, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36,
	38, 24, 21, 2, 3, 5, 6, 1,
	9, 10, 11, 12, 13, 14, 15, 16,
	17, 18, 19, 20, 1, 48, 1, 47,
	1, 1, 1, 60,
}

var _myaSM_trans_targs []byte = []byte{
	0, 1, 26, 37, 0, 27, 29, 51,
	54, 39, 40, 41, 28, 43, 44, 46,
	47, 48, 30, 50, 45, 0, 2, 13,
	0, 3, 5, 14, 15, 16, 4, 18,
	19, 21, 22, 23, 6, 25, 20, 12,
	9, 10, 11, 7, 8, 17, 24, 0,
	0, 36, 33, 34, 35, 31, 32, 38,
	42, 49, 52, 53, 0,
}

var _myaSM_trans_actions []byte = []byte{
	11, 0, 0, 0, 7, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 13, 0, 0,
	5, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 15,
	9, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 17,
}

var _myaSM_to_state_actions []byte = []byte{
	1, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0,
}

var _myaSM_from_state_actions []byte = []byte{
	3, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0,
}

var _myaSM_eof_trans []int16 = []int16{
	0, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 48, 48, 48, 48, 48, 48,
	48, 48, 48, 48, 48, 48, 48, 48,
	48, 48, 48, 48, 48, 48, 48, 48,
	48, 48, 48, 22, 22, 48, 61,
}

const myaSM_start int = 0
const myaSM_first_final int = 0
const myaSM_error int = -1

const myaSM_en_main int = 0

func findSyllablesMyanmar(buffer *Buffer) {
	var p, ts, te, act, cs int
	info := buffer.Info

	{
		cs = myaSM_start
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
		_acts = int(_myaSM_from_state_actions[cs])
		_nacts = uint(_myaSM_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _myaSM_actions[_acts-1] {
			case 1:
				ts = p

			}
		}

		_keys = int(_myaSM_key_offsets[cs])
		_trans = int(_myaSM_index_offsets[cs])

		_klen = int(_myaSM_single_lengths[cs])
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
				case (info[p].complexCategory) < _myaSM_trans_keys[_mid]:
					_upper = _mid - 1
				case (info[p].complexCategory) > _myaSM_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_myaSM_range_lengths[cs])
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
				case (info[p].complexCategory) < _myaSM_trans_keys[_mid]:
					_upper = _mid - 2
				case (info[p].complexCategory) > _myaSM_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_myaSM_indicies[_trans])
	_eof_trans:
		cs = int(_myaSM_trans_targs[_trans])

		if _myaSM_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_myaSM_trans_actions[_trans])
		_nacts = uint(_myaSM_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _myaSM_actions[_acts-1] {
			case 2:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarConsonantSyllable, ts, te, info, &syllableSerial)
				}
			case 3:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarNonMyanmarCluster, ts, te, info, &syllableSerial)
				}
			case 4:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarBrokenCluster, ts, te, info, &syllableSerial)
					buffer.scratchFlags |= bsfHasBrokenSyllable
				}
			case 5:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarNonMyanmarCluster, ts, te, info, &syllableSerial)
				}
			case 6:
				te = p
				p--
				{
					foundSyllableMyanmar(myanmarConsonantSyllable, ts, te, info, &syllableSerial)
				}
			case 7:
				te = p
				p--
				{
					foundSyllableMyanmar(myanmarBrokenCluster, ts, te, info, &syllableSerial)
					buffer.scratchFlags |= bsfHasBrokenSyllable
				}
			case 8:
				te = p
				p--
				{
					foundSyllableMyanmar(myanmarNonMyanmarCluster, ts, te, info, &syllableSerial)
				}
			}
		}

	_again:
		_acts = int(_myaSM_to_state_actions[cs])
		_nacts = uint(_myaSM_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _myaSM_actions[_acts-1] {
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
			if _myaSM_eof_trans[cs] > 0 {
				_trans = int(_myaSM_eof_trans[cs] - 1)
				goto _eof_trans
			}
		}

	}

	_ = act // needed by Ragel, but unused
}
