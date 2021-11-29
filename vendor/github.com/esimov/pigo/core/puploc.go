package pigo

import (
	"encoding/binary"
	"math"
	"math/rand"
	"sort"
	"sync"
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

// NewPuplocCascade initializes the PuplocCascade constructor method.
func NewPuplocCascade() *PuplocCascade {
	return &PuplocCascade{}
}

// UnpackCascade unpacks the pupil localization cascade file.
func (plc *PuplocCascade) UnpackCascade(packet []byte) (*PuplocCascade, error) {
	var (
		stages    uint32
		scales    float32
		trees     uint32
		treeDepth uint32

		treeCodes = make([]int8, 0, 409200)
		treePreds = make([]float32, 0, 204800)
	)

	pos := 0
	// Get the number of stages as 32-bit unsigned integer.
	stages = binary.LittleEndian.Uint32(packet[pos:])
	pos += 4

	// Obtain the scale multiplier (applied after each stage).
	u32scales := binary.LittleEndian.Uint32(packet[pos:])
	// Convert uint32 to float32
	scales = *(*float32)(unsafe.Pointer(&u32scales))
	pos += 4

	// Obtain the number of trees per stage.
	trees = binary.LittleEndian.Uint32(packet[pos:])
	pos += 4

	// Obtain the depth of each tree.
	treeDepth = binary.LittleEndian.Uint32(packet[pos:])
	pos += 4

	// Traverse all the stages of the binary tree.
	for s := 0; s < int(stages); s++ {
		// Traverse the branches of each stage.
		for t := 0; t < int(trees); t++ {
			depth := int(pow(2, int(treeDepth)))

			code := packet[pos : pos+4*depth-4]
			// Convert unsigned bytecodes to signed ones.
			i8code := *(*[]int8)(unsafe.Pointer(&code))
			treeCodes = append(treeCodes, i8code...)

			pos += 4*depth - 4

			// Read prediction from tree's leaf nodes.
			for i := 0; i < depth; i++ {
				for l := 0; l < 2; l++ {
					u32pred := binary.LittleEndian.Uint32(packet[pos:])
					// Convert uint32 to float32
					f32pred := *(*float32)(unsafe.Pointer(&u32pred))
					treePreds = append(treePreds, f32pred)
					pos += 4
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

// classifyRegion applies the face classification function over an image.
func (plc *PuplocCascade) classifyRegion(r, c, s float32, treeDepth, nrows, ncols int, pixels []uint8, dim int, flipV bool) []float32 {
	var (
		c1, c2 int
		root   int
	)

	for i := 0; i < int(plc.stages); i++ {
		var dr, dc float32 = 0.0, 0.0

		for j := 0; j < int(plc.trees); j++ {
			idx := 0
			for k := 0; k < int(plc.treeDepth); k++ {
				r1 := min(nrows-1, max(0, (256*int(r)+int(plc.treeCodes[root+4*idx+0])*int(math.Round(float64(s))))>>8))
				r2 := min(nrows-1, max(0, (256*int(r)+int(plc.treeCodes[root+4*idx+2])*int(math.Round(float64(s))))>>8))

				// flipV means that we wish to flip the column coordinates sign in the tree nodes.
				// This is required at running the facial landmark detector over the right side of the detected face.
				if flipV {
					c1 = min(ncols-1, max(0, (256*int(c)+int(-plc.treeCodes[root+4*idx+1])*int(math.Round(float64(s))))>>8))
					c2 = min(ncols-1, max(0, (256*int(c)+int(-plc.treeCodes[root+4*idx+3])*int(math.Round(float64(s))))>>8))
				} else {
					c1 = min(ncols-1, max(0, (256*int(c)+int(plc.treeCodes[root+4*idx+1])*int(math.Round(float64(s))))>>8))
					c2 = min(ncols-1, max(0, (256*int(c)+int(plc.treeCodes[root+4*idx+3])*int(math.Round(float64(s))))>>8))
				}
				bintest := func(p1, p2 uint8) uint8 {
					if p1 > p2 {
						return 1
					}
					return 0
				}
				idx = 2*idx + 1 + int(bintest(pixels[r1*dim+c1], pixels[r2*dim+c2]))
			}
			lutIdx := 2 * (int(plc.trees)*treeDepth*i + treeDepth*j + idx - (treeDepth - 1))

			dr += plc.treePreds[lutIdx+0]
			if flipV {
				dc += -plc.treePreds[lutIdx+1]
			} else {
				dc += plc.treePreds[lutIdx+1]
			}
			root += 4*treeDepth - 4
		}

		r += dr * s
		c += dc * s
		s *= plc.scales
	}
	return []float32{r, c, s}
}

// classifyRotatedRegion applies the face classification function over a rotated image.
func (plc *PuplocCascade) classifyRotatedRegion(r, c, s float32, a float64, treeDepth, nrows, ncols int, pixels []uint8, dim int, flipV bool) []float32 {
	var (
		row1, col1, row2, col2 int
		root                   int
	)

	qCosTable := []float32{256, 251, 236, 212, 181, 142, 97, 49, 0, -49, -97, -142, -181, -212, -236, -251, -256, -251, -236, -212, -181, -142, -97, -49, 0, 49, 97, 142, 181, 212, 236, 251, 256}
	qSinTable := []float32{0, 49, 97, 142, 181, 212, 236, 251, 256, 251, 236, 212, 181, 142, 97, 49, 0, -49, -97, -142, -181, -212, -236, -251, -256, -251, -236, -212, -181, -142, -97, -49, 0}

	qsin := s * qSinTable[int(32.0*a)] //s*(256.0*math.Sin(2*math.Pi*a))
	qcos := s * qCosTable[int(32.0*a)] //s*(256.0*math.Cos(2*math.Pi*a))

	for i := 0; i < int(plc.stages); i++ {
		var dr, dc float32 = 0.0, 0.0

		for j := 0; j < int(plc.trees); j++ {
			idx := 0
			for k := 0; k < int(plc.treeDepth); k++ {
				row1 = int(plc.treeCodes[root+4*idx+0])
				row2 = int(plc.treeCodes[root+4*idx+2])

				// flipV means that we wish to flip the column coordinates sign in the tree nodes.
				// This is required at running the facial landmark detector over the right side of the detected face.
				if flipV {
					col1 = int(-plc.treeCodes[root+4*idx+1])
					col2 = int(-plc.treeCodes[root+4*idx+3])
				} else {
					col1 = int(plc.treeCodes[root+4*idx+1])
					col2 = int(plc.treeCodes[root+4*idx+3])
				}

				r1 := min(nrows-1, max(0, 65536*int(r)+int(qcos)*row1-int(qsin)*col1)>>16)
				c1 := min(ncols-1, max(0, 65536*int(c)+int(qsin)*row1+int(qcos)*col1)>>16)
				r2 := min(nrows-1, max(0, 65536*int(r)+int(qcos)*row2-int(qsin)*col2)>>16)
				c2 := min(ncols-1, max(0, 65536*int(c)+int(qsin)*row2+int(qcos)*col2)>>16)

				bintest := func(px1, px2 uint8) int {
					if px1 <= px2 {
						return 1
					}
					return 0
				}
				idx = 2*idx + 1 + bintest(pixels[r1*dim+c1], pixels[r2*dim+c2])
			}
			lutIdx := 2 * (int(plc.trees)*treeDepth*i + treeDepth*j + idx - (treeDepth - 1))

			dr += plc.treePreds[lutIdx+0]
			if flipV {
				dc += -plc.treePreds[lutIdx+1]
			} else {
				dc += plc.treePreds[lutIdx+1]
			}
			root += 4*treeDepth - 4
		}

		r += dr * s
		c += dc * s
		s *= plc.scales
	}
	return []float32{r, c, s}
}

// puplocPool is a struct for holding the pupil localization values in sync pool.
type puplocPool struct {
	rows  []float32
	cols  []float32
	scale []float32
}

// Create a sync.Pool for further reusing the allocated memory space
// in order to keep the GC overhead as low as possible.
var plcPool = sync.Pool{
	New: func() interface{} {
		return &puplocPool{
			rows:  make([]float32, 63),
			cols:  make([]float32, 63),
			scale: make([]float32, 63),
		}
	},
}

// RunDetector runs the pupil localization function.
func (plc *PuplocCascade) RunDetector(pl Puploc, img ImageParams, angle float64, flipV bool) *Puploc {
	var res = make([]float32, 3)

	det := plcPool.Get().(*puplocPool)
	defer plcPool.Put(det)

	treeDepth := int(pow(2, int(plc.treeDepth)))

	for i := 0; i < pl.Perturbs; i++ {
		row := float32(pl.Row) + float32(pl.Scale)*0.15*(0.5-rand.Float32())
		col := float32(pl.Col) + float32(pl.Scale)*0.15*(0.5-rand.Float32())
		sc := float32(pl.Scale) * (0.925 + 0.15*rand.Float32())

		if angle > 0.0 {
			if angle > 1.0 {
				angle = 1.0
			}
			res = plc.classifyRotatedRegion(row, col, sc, angle, treeDepth, img.Rows, img.Cols, img.Pixels, img.Dim, flipV)
		} else {
			res = plc.classifyRegion(row, col, sc, treeDepth, img.Rows, img.Cols, img.Pixels, img.Dim, flipV)
		}

		det.rows[i] = res[0]
		det.cols[i] = res[1]
		det.scale[i] = res[2]
	}

	// Sorting the perturbations in ascendent order
	sort.Sort(plocSort(det.rows))
	sort.Sort(plocSort(det.cols))
	sort.Sort(plocSort(det.scale))

	// Get the median value of the sorted perturbation results
	return &Puploc{
		Row:   int(det.rows[int(math.Round(float64(pl.Perturbs)/2))]),
		Col:   int(det.cols[int(math.Round(float64(pl.Perturbs)/2))]),
		Scale: det.scale[int(math.Round(float64(pl.Perturbs)/2))],
	}
}

// Implement custom sorting function on detection values.
type plocSort []float32

func (q plocSort) Len() int           { return len(q) }
func (q plocSort) Less(i, j int) bool { return q[i] < q[j] }
func (q plocSort) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
