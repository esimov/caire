package harfbuzz 

// Code generated with ragel -Z -o ot_indic_machine.go ot_indic_machine.rl ; sed -i '/^\/\/line/ d' ot_indic_machine.go ; goimports -w ot_indic_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-indic-machine.rl Copyright © 2015 Google, Inc. Behdad Esfahbod

// indic_syllable_type_t
const  (
  indicConsonantSyllable = iota
  indicVowelSyllable
  indicStandaloneCluster
  indicSymbolCluster
  indicBrokenCluster
  indicNonIndicCluster
)

%%{
  machine indSM;
  alphtype byte;
  write exports;
  write data;
}%%

%%{


export X    = 0;
export C    = 1;
export V    = 2;
export N    = 3;
export H    = 4;
export ZWNJ = 5;
export ZWJ  = 6;
export M    = 7;
export SM   = 8;
export A    = 9;
export VD   = 9;
export PLACEHOLDER = 10;
export DOTTEDCIRCLE = 11;
export RS    = 12;
export MPst  = 13;
export Repha = 14;
export Ra    = 15;
export CM    = 16;
export Symbol= 17;
export CS    = 18;

c = (C | Ra);			# is_consonant
n = ((ZWNJ?.RS)? (N.N?)?);	# is_consonant_modifier
z = ZWJ|ZWNJ;			# is_joiner
reph = (Ra H | Repha);		# possible reph

cn = c.ZWJ?.n?;
symbol = Symbol.N?;
matra_group = z*.(M | SM? MPst).N?.H?;
syllable_tail = (z?.SM.SM?.ZWNJ?)? (A | VD)*;
halant_group = (z?.H.(ZWJ.N?)?);
final_halant_group = halant_group | H.ZWNJ;
medial_group = CM?;
halant_or_matra_group = (final_halant_group | matra_group*);

complex_syllable_tail = (halant_group.cn)* medial_group halant_or_matra_group syllable_tail;

consonant_syllable =	(Repha|CS)? cn complex_syllable_tail;
vowel_syllable =	reph? V.n? (ZWJ | complex_syllable_tail);
standalone_cluster =	((Repha|CS)? PLACEHOLDER | reph? DOTTEDCIRCLE).n? complex_syllable_tail;
symbol_cluster =	symbol syllable_tail;
broken_cluster =	reph? n? complex_syllable_tail;
other =			any;

main := |*
	consonant_syllable	=> { foundSyllableIndic (indicConsonantSyllable,ts, te, info, &syllableSerial); };
	vowel_syllable		=> { foundSyllableIndic (indicVowelSyllable,ts, te, info, &syllableSerial); };
	standalone_cluster	=> { foundSyllableIndic (indicStandaloneCluster,ts, te, info, &syllableSerial); };
	symbol_cluster		=> { foundSyllableIndic (indicSymbolCluster,ts, te, info, &syllableSerial); };
	broken_cluster		=> { foundSyllableIndic (indicBrokenCluster,ts, te, info, &syllableSerial); buffer.scratchFlags |= bsfHasBrokenSyllable; };
	other			=> { foundSyllableIndic (indicNonIndicCluster,ts, te, info, &syllableSerial); };
*|;

}%%

func findSyllablesIndic (buffer * Buffer) {
    var p, ts, te, act, cs int 
    info := buffer.Info;
    %%{
        write init;
        getkey info[p].complexCategory;
    }%%

    pe := len(info)
    eof := pe

    var syllableSerial uint8 = 1;
    %%{
        write exec;
    }%%
    _ = act // needed by Ragel, but unused
}

