package harfbuzz 

// Code generated with ragel -Z -o ot_myanmar_machine.go ot_myanmar_machine.rl ; sed -i '/^\/\/line/ d' ot_myanmar_machine.go ; goimports -w ot_myanmar_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-myanmar-machine.rl Copyright © 2015 Mozilla Foundation. Google, Inc. Behdad Esfahbod

// myanmar_syllable_type_t
const  (
  myanmarConsonantSyllable = iota
  myanmarBrokenCluster
  myanmarNonMyanmarCluster
)

%%{
  machine myaSM;
  alphtype byte;
  write exports;
  write data;
}%%

%%{

# Spec category D is folded into GB; D0 is not implemented by Uniscribe and as such folded into D
# Spec category P is folded into GB

export C    = 1;
export IV   = 2;
export DB   = 3;	# Dot below	     = OT_N
export H    = 4;
export ZWNJ = 5;
export ZWJ  = 6;
export SM    = 8;	# Visarga and Shan tones
export GB   = 10;	# 		     = OT_PLACEHOLDER
export DOTTEDCIRCLE = 11;
export A    = 9;
export Ra   = 15;
export CS   = 18;

export VAbv = 20;
export VBlw = 21;
export VPre = 22;
export VPst = 23;

# 32+ are for Myanmar-specific values
export As   = 32;	# Asat
export MH   = 35;	# Medial Ha
export MR   = 36;	# Medial Ra
export MW   = 37;	# Medial Wa, Shan Wa
export MY   = 38;	# Medial Ya, Mon Na, Mon Ma
export PT   = 39;	# Pwo and other tones
export VS   = 40;	# Variation selectors
export ML   = 41;	# Medial Mon La

j = ZWJ|ZWNJ;			# Joiners
k = (Ra As H);			# Kinzi

c = C|Ra;			# is_consonant

medial_group = MY? As? MR? ((MW MH? ML? | MH ML? | ML) As?)?;
main_vowel_group = (VPre.VS?)* VAbv* VBlw* A* (DB As?)?;
post_vowel_group = VPst MH? ML? As* VAbv* A* (DB As?)?;
pwo_tone_group = PT A* DB? As?;

complex_syllable_tail = As* medial_group main_vowel_group post_vowel_group* pwo_tone_group* SM* j?;
syllable_tail = (H (c|IV).VS?)* (H | complex_syllable_tail);

consonant_syllable =	(k|CS)? (c|IV|GB|DOTTEDCIRCLE).VS? syllable_tail;
broken_cluster =	k? VS? syllable_tail;
other =			any;

main := |*
	consonant_syllable	=> { foundSyllableMyanmar (myanmarConsonantSyllable, ts, te, info, &syllableSerial); };
	j			=> { foundSyllableMyanmar (myanmarNonMyanmarCluster, ts, te, info, &syllableSerial); };
	broken_cluster		=> { foundSyllableMyanmar (myanmarBrokenCluster, ts, te, info, &syllableSerial); buffer.scratchFlags |= bsfHasBrokenSyllable };
	other			=> { foundSyllableMyanmar (myanmarNonMyanmarCluster, ts, te, info, &syllableSerial); };
*|;


}%%


func findSyllablesMyanmar (buffer *Buffer){
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

