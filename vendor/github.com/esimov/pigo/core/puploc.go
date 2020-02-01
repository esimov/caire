package pigo

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"sort"
	"unsafe"
)

// Puploc contains all the information resulted from the pupil detection
// needed for accessing from a global scope.
type Puploc struct {
	Row      int
	Col      int
	Scale    float32
	Perturbs int
}

// PuplocCascade is a general struct for storing
// the cascade tree values encoded into the binary file.
type PuplocCascade struct {
	stages    uint32
	scales    float32
	trees     uint32
	treeDepth uint32
	treeCodes []int8
	treePreds []float32
}

// UnpackCascade unpacks the pupil localization cascade file.
func (plc *PuplocCascade) UnpackCascade(packet []byte) (*PuplocCascade, error) {
	var (
		stages    uint32
		scales    float32
		trees     uint32
		treeDepth uint32
		treeCodes []int8
		treePreds []float32
	)

	pos := 0
	buff := make([]byte, 4)
	dataView := bytes.NewBuffer(buff)

	// Read the depth (size) of each tree and write it into the buffer array.
	_, err := dataView.Write([]byte{packet[pos+0], packet[pos+1], packet[pos+2], packet[pos+3]})
	if err != nil {
		return nil, err
	}

	if dataView.Len() > 0 {
		// Get the number of stages as 32-bit uint and write it into the buffer array.
		stages = binary.LittleEndian.Uint32(packet[pos:])
		_, err := dataView.Write([]byte{packet[pos+0], packet[pos+1], packet[pos+2], packet[pos+3]})
		if err != nil {
			return nil, err
		}
		pos += 4

		// Obtain the scale multiplier (applied after each stage) and write it into the buffer array.
		u32scales := binary.LittleEndian.Uint32(packet[pos:])
		// Convert uint32 to float32
		scales = *(*float32)(unsafe.Pointer(&u32scales))
		_, err = dataView.Write([]byte{packet[pos+0], packet[pos+1], packet[pos+2], packet[pos+3]})
		if err != nil {
			return nil, err
		}
		pos += 4

		// Obtain the number of trees per stage and write it into the buffer array.
		trees = binary.LittleEndian.Uint32(packet[pos:])
		_, err = dataView.Write([]byte{packet[pos+0], packet[pos+1], packet[pos+2], packet[pos+3]})
		if err != nil {
			return nil, err
		}
		pos += 4

		// Obtain the depth of each tree and write it into the buffer array.
		treeDepth = binary.LittleEndian.Uint32(packet[pos:])
		_, err = dataView.Write([]byte{packet[pos+0], packet[pos+1], packet[pos+2], packet[pos+3]})
		if err != nil {
			return nil, err
		}
		pos += 4

		// Traverse all the stages of the binary tree
		for s := 0; s < int(stages); s++ {
			// Traverse the branches of each stage
			for t := 0; t < int(trees); t++ {
				code := packet[pos : pos+int(4*math.Pow(2, float64(treeDepth))-4)]
				// Convert unsigned bytecodes to signed ones.
				i8code := *(*[]int8)(unsafe.Pointer(&code))
				treeCodes = append(treeCodes, i8code...)

				pos = pos + int(4*math.Pow(2, float64(treeDepth))-4)

				// Read prediction from tree's leaf nodes.
				for i := 0; i < int(math.Pow(2, float64(treeDepth))); i++ {
					for l := 0; l < 2; l++ {
						_, err := dataView.Write([]byte{packet[pos+0], packet[pos+1], packet[pos+2], packet[pos+3]})
						if err != nil {
							return nil, err
						}
						u32pred := binary.LittleEndian.Uint32(packet[pos:])
						// Convert uint32 to float32
						f32pred := *(*float32)(unsafe.Pointer(&u32pred))
						treePreds = append(treePreds, f32pred)
						pos += 4
					}
				}
			}

		}
	}

	return &PuplocCascade{
		stages:    stages,
		scales:    scales,
		trees:     trees,
		treeDepth: treeDepth,
		treeCodes: treeCodes,
		treePreds: treePreds,
	}, nil
}

// RunDetector runs the pupil localization function.
func (plc *PuplocCascade) RunDetector(pl Puploc, img ImageParams) *Puploc {
	localization := func(r, c, s float32, pixels []uint8, rows, cols, dim int) []float32 {
		root := 0
		pTree := int(math.Pow(2, float64(plc.treeDepth)))

		for i := 0; i < int(plc.stages); i++ {
			var dr, dc float32 = 0.0, 0.0

			for j := 0; j < int(plc.trees); j++ {
				idx := 0
				for k := 0; k < int(plc.treeDepth); k++ {
					r1 := min(rows-1, max(0, (256*int(r)+int(plc.treeCodes[root+4*idx+0])*int(round(float64(s))))>>8))
					c1 := min(cols-1, max(0, (256*int(c)+int(plc.treeCodes[root+4*idx+1])*int(round(float64(s))))>>8))
					r2 := min(rows-1, max(0, (256*int(r)+int(plc.treeCodes[root+4*idx+2])*int(round(float64(s))))>>8))
					c2 := min(cols-1, max(0, (256*int(c)+int(plc.treeCodes[root+4*idx+3])*int(round(float64(s))))>>8))

					bintest := func(r1, r2 uint8) uint8 {
						if r1 > r2 {
							return 1
						}
						return 0
					}
					idx = 2*idx + 1 + int(bintest(pixels[r1*dim+c1], pixels[r2*dim+c2]))
				}
				lutIdx := 2 * (int(plc.trees)*pTree*i + pTree*j + idx - (pTree - 1))

				dr += plc.treePreds[lutIdx+0]
				dc += plc.treePreds[lutIdx+1]

				root += 4*pTree - 4
			}

			r += dr * s
			c += dc * s
			s *= plc.scales
		}
		return []float32{r, c, s}
	}
	rows, cols, scale := []float32{}, []float32{}, []float32{}

	for i := 0; i < pl.Perturbs; i++ {
		rt := float32(pl.Row) + float32(pl.Scale)*0.15*(0.5-rand.Float32())
		ct := float32(pl.Col) + float32(pl.Scale)*0.15*(0.5-rand.Float32())
		st := float32(pl.Scale) * (0.25 + rand.Float32())

		res := localization(rt, ct, st, img.Pixels, img.Rows, img.Cols, img.Dim)

		rows = append(rows, res[0])
		cols = append(cols, res[1])
		scale = append(scale, res[2])
	}

	// sorting the perturbations in ascendent order
	sort.Sort(plocSort(rows))
	sort.Sort(plocSort(cols))
	sort.Sort(plocSort(scale))

	// get the median value of the sorted perturbation results
	return &Puploc{
		Row:   int(rows[int(round(float64(pl.Perturbs)/2))]),
		Col:   int(cols[int(round(float64(pl.Perturbs)/2))]),
		Scale: scale[int(round(float64(pl.Perturbs)/2))],
	}
}

// min returns the minum value between two numbers
func min(val1, val2 int) int {
	if val1 < val2 {
		return val1
	}
	return val2
}

// max returns the maximum value between two numbers
func max(val1, val2 int) int {
	if val1 > val2 {
		return val1
	}
	return val2
}

// round returns the nearest integer, rounding ties away from zero.
func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

// Implement custom sorting function on detection values.
type plocSort []float32

func (q plocSort) Len() int      { return len(q) }
func (q plocSort) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q plocSort) Less(i, j int) bool {
	if q[i] < q[j] {
		return true
	}
	if q[i] > q[j] {
		return false
	}
	return q[i] < q[j]
}
