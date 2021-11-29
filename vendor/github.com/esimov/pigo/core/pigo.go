package pigo

import (
	"encoding/binary"
	"math"
	"sort"
	"sync"
	"unsafe"
)

// CascadeParams contains the basic parameters to run the analyzer function over the defined image.
// MinSize: represents the minimum size of the face.
// MaxSize: represents the maximum size of the face.
// ShiftFactor: determines to what percentage to move the detection window over its size.
// ScaleFactor: defines in percentage the resize value of the detection window when moving to a higher scale.
type CascadeParams struct {
	MinSize     int
	MaxSize     int
	ShiftFactor float64
	ScaleFactor float64
	ImageParams
}

// ImageParams is a struct for image related settings.
// Pixels: contains the grayscale converted image pixel data.
// Rows: the number of image rows.
// Cols: the number of image columns.
// Dim: the image dimension.
type ImageParams struct {
	Pixels []uint8
	Rows   int
	Cols   int
	Dim    int
}

// Pigo struct defines the basic binary tree components.
type Pigo struct {
	treeDepth     uint32
	treeNum       uint32
	treeCodes     []int8
	treePred      []float32
	treeThreshold []float32
}

// NewPigo initializes the Pigo constructor method.
func NewPigo() *Pigo {
	return &Pigo{}
}

// Unpack unpack the binary face classification file.
func (pg *Pigo) Unpack(packet []byte) (*Pigo, error) {
	var (
		treeDepth     uint32
		treeNum       uint32
		treeCodes     []int8
		treePred      []float32
		treeThreshold []float32
	)

	// We skip the first 8 bytes of the cascade file.
	pos := 8

	// Obtain the depth of each tree from the binary data.
	treeDepth = binary.LittleEndian.Uint32(packet[pos:])
	pos += 4

	// Get the number of cascade trees as 32-bit unsigned integer.
	treeNum = binary.LittleEndian.Uint32(packet[pos:])

	// To avoid constant memory allocation on each append we predefine the slice capacity.
	treeThreshold = make([]float32, 0, treeNum)
	treeCodes = make([]int8, 0, 119808)
	treePred = make([]float32, 0, 29952)

	pos += 4

	for t := 0; t < int(treeNum); t++ {
		// Obtain the tree codes of each tree nodes.
		treeCodes = append(treeCodes, []int8{0, 0, 0, 0}...)

		code := packet[pos : pos+int(4*pow(2, int(treeDepth))-4)]
		// Convert unsigned bytecodes to signed ones.
		signedCode := *(*[]int8)(unsafe.Pointer(&code))
		treeCodes = append(treeCodes, signedCode...)

		pos += int(4*pow(2, int(treeDepth)) - 4)

		// Read prediction from tree's leaf nodes.
		for i := 0; i < int(pow(2, int(treeDepth))); i++ {
			u32pred := binary.LittleEndian.Uint32(packet[pos:])
			// Convert uint32 to float32
			f32pred := *(*float32)(unsafe.Pointer(&u32pred))
			treePred = append(treePred, f32pred)
			pos += 4
		}
		u32thr := binary.LittleEndian.Uint32(packet[pos:])
		// Convert uint32 to float32
		f32thr := *(*float32)(unsafe.Pointer(&u32thr))
		treeThreshold = append(treeThreshold, f32thr)
		pos += 4
	}

	return &Pigo{
		treeDepth,
		treeNum,
		treeCodes,
		treePred,
		treeThreshold,
	}, nil
}

// classifyRegion constructs the classification function based on the parsed binary data.
func (pg *Pigo) classifyRegion(r, c, s, treeDepth int, pixels []uint8, dim int) float32 {
	var (
		root int
		out  float32
	)

	r = r * 256
	c = c * 256

	if pg.treeNum > 0 {
		for i := 0; i < int(pg.treeNum); i++ {
			idx := 1
			for j := 0; j < int(pg.treeDepth); j++ {
				x1 := ((r+int(pg.treeCodes[root+4*idx+0])*s)>>8)*dim + ((c + int(pg.treeCodes[root+4*idx+1])*s) >> 8)
				x2 := ((r+int(pg.treeCodes[root+4*idx+2])*s)>>8)*dim + ((c + int(pg.treeCodes[root+4*idx+3])*s) >> 8)

				bintest := func(px1, px2 uint8) int {
					if px1 <= px2 {
						return 1
					}
					return 0
				}
				idx = 2*idx + bintest(pixels[x1], pixels[x2])
			}
			out += pg.treePred[treeDepth*i+idx-treeDepth]

			if out <= pg.treeThreshold[i] {
				return -1.0
			}
			root += 4 * treeDepth
		}
		return out - pg.treeThreshold[pg.treeNum-1]
	}
	return 0.0
}

// classifyRotatedRegion applies the face classification function over a rotated image based on the parsed binary data.
func (pg *Pigo) classifyRotatedRegion(r, c, s, treeDepth int, a float64, nrows, ncols int, pixels []uint8, dim int) float32 {
	var (
		root int
		out  float32
	)

	qCosTable := []int{256, 251, 236, 212, 181, 142, 97, 49, 0, -49, -97, -142, -181, -212, -236, -251, -256, -251, -236, -212, -181, -142, -97, -49, 0, 49, 97, 142, 181, 212, 236, 251, 256}
	qSinTable := []int{0, 49, 97, 142, 181, 212, 236, 251, 256, 251, 236, 212, 181, 142, 97, 49, 0, -49, -97, -142, -181, -212, -236, -251, -256, -251, -236, -212, -181, -142, -97, -49, 0}

	qsin := s * qSinTable[int(32.0*a)] //s*(256.0*math.Sin(2*math.Pi*a))
	qcos := s * qCosTable[int(32.0*a)] //s*(256.0*math.Cos(2*math.Pi*a))

	if pg.treeNum > 0 {
		for i := 0; i < int(pg.treeNum); i++ {
			var idx = 1

			for j := 0; j < int(pg.treeDepth); j++ {
				r1 := abs(min(nrows-1, max(0, 65536*r+qcos*int(pg.treeCodes[root+4*idx+0])-qsin*int(pg.treeCodes[root+4*idx+1]))>>16))
				c1 := abs(min(nrows-1, max(0, 65536*c+qsin*int(pg.treeCodes[root+4*idx+0])+qcos*int(pg.treeCodes[root+4*idx+1]))>>16))

				r2 := abs(min(nrows-1, max(0, 65536*r+qcos*int(pg.treeCodes[root+4*idx+2])-qsin*int(pg.treeCodes[root+4*idx+3]))>>16))
				c2 := abs(min(nrows-1, max(0, 65536*c+qsin*int(pg.treeCodes[root+4*idx+2])+qcos*int(pg.treeCodes[root+4*idx+3]))>>16))

				bintest := func(px1, px2 uint8) int {
					if px1 <= px2 {
						return 1
					}
					return 0
				}
				idx = 2*idx + bintest(pixels[r1*dim+c1], pixels[r2*dim+c2])
			}
			out += pg.treePred[treeDepth*i+idx-treeDepth]

			if out <= pg.treeThreshold[i] {
				return -1.0
			}
			root += 4 * treeDepth
		}
		return out - pg.treeThreshold[pg.treeNum-1]
	}
	return 0.0
}

// Detection struct contains the detection results composed of
// the row, column, scale factor and the detection score.
type Detection struct {
	Row   int
	Col   int
	Scale int
	Q     float32
}

// We are using sync.Pool to avoid memory allocation on the heap
// in order to keep the GC overhead as small as possible.
var detpool = sync.Pool{
	New: func() interface{} {
		return &Detection{}
	},
}

// RunCascade analyze the grayscale converted image pixel data and run the classification function over the detection window.
// It will return a slice containing the detection row, column, it's center and the detection score (in case this is greater than 0.0).
func (pg *Pigo) RunCascade(cp CascadeParams, angle float64) []Detection {
	var (
		detections []Detection
		pixels     = cp.Pixels
		treeDepth  = int(pow(2, int(pg.treeDepth)))
		q          float32
	)
	scale := cp.MinSize

	det := detpool.Get().(*Detection)
	defer detpool.Put(det)

	// Run the classification function over the detection window
	// and check if the false positive rate is above a certain value.
	for scale <= cp.MaxSize {
		step := int(math.Max(cp.ShiftFactor*float64(scale), 1))
		offset := (scale/2 + 1)

		for row := offset; row <= cp.Rows-offset; row += step {
			for col := offset; col <= cp.Cols-offset; col += step {
				if angle > 0.0 {
					if angle > 1.0 {
						angle = 1.0
					}
					q = pg.classifyRotatedRegion(row, col, scale, treeDepth, angle, cp.Rows, cp.Cols, pixels, cp.Dim)
				} else {
					q = pg.classifyRegion(row, col, scale, treeDepth, pixels, cp.Dim)
				}

				det.Row = row
				det.Col = col
				det.Scale = scale
				det.Q = q

				if q > 0.0 {
					detections = append(detections, *det)
				}
			}
		}
		// We need to avoid running into an infinite loop because of float to int conversion
		// in cases when scaleFactor == 1.1 and minSize == 9 as example.
		// When the scale is 9, the factor would come up with 9.9, which again becomes 9 because of the int() conversion.
		// This approach gives the same speed without having an impact on the detection score.
		scale = int(float64(scale) + math.Max(2, (float64(scale)*cp.ScaleFactor)-float64(scale)))
	}
	return detections
}

// ClusterDetections returns the intersection over union of multiple clusters.
// We need to make this comparison to filter out multiple face detection regions.
func (pg *Pigo) ClusterDetections(detections []Detection, iouThreshold float64) []Detection {
	// Sort detections by their score
	sort.Sort(det(detections))

	calcIoU := func(det1, det2 Detection) float64 {
		// Unpack the position and size of each detection.
		r1, c1, s1 := float64(det1.Row), float64(det1.Col), float64(det1.Scale)
		r2, c2, s2 := float64(det2.Row), float64(det2.Col), float64(det2.Scale)

		overRow := math.Max(0, math.Min(r1+s1/2, r2+s2/2)-math.Max(r1-s1/2, r2-s2/2))
		overCol := math.Max(0, math.Min(c1+s1/2, c2+s2/2)-math.Max(c1-s1/2, c2-s2/2))

		// Return intersection over union.
		return overRow * overCol / (s1*s1 + s2*s2 - overRow*overCol)
	}
	assignments := make([]bool, len(detections))
	clusters := []Detection{}

	for i := 0; i < len(detections); i++ {
		// Compare the intersection over union only for two different clusters.
		// Skip the comparison in case there already exists a cluster A in the bucket.
		if !assignments[i] {
			var (
				r, c, s, n int
				q          float32
			)
			for j := 0; j < len(detections); j++ {
				// Check if the comparison result is above a certain threshold.
				// In this case we union the detections.
				if calcIoU(detections[i], detections[j]) > iouThreshold {
					assignments[j] = true
					r += detections[j].Row
					c += detections[j].Col
					s += detections[j].Scale
					q += detections[j].Q
					n++
				}
			}
			if n > 0 {
				clusters = append(clusters, Detection{r / n, c / n, s / n, q})
			}
		}
	}
	return clusters
}

// Implement sorting function on detection values.
type det []Detection

func (q det) Len() int      { return len(q) }
func (q det) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q det) Less(i, j int) bool {
	return q[i].Q < q[j].Q
}
