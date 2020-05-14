package pigo

import (
	"errors"
	"io/ioutil"
	"math"
	"path/filepath"
)

// FlpCascade holds the binary representation of the facial landmark points cascade files
type FlpCascade struct {
	*PuplocCascade
	error
}

// UnpackFlp unpacks the facial landmark points cascade file.
// This will return the binary representation of the cascade file.
func (plc *PuplocCascade) UnpackFlp(cf string) (*PuplocCascade, error) {
	flpc, err := ioutil.ReadFile(cf)
	if err != nil {
		return nil, err
	}
	return plc.UnpackCascade(flpc)
}

// FindLandmarkPoints detects the facial landmark points based on the pupil localization results.
func (plc *PuplocCascade) FindLandmarkPoints(leftEye, rightEye *Puploc, img ImageParams, perturb int, flipV bool) *Puploc {
	var flploc *Puploc
	dist1 := (leftEye.Row - rightEye.Row) * (leftEye.Row - rightEye.Row)
	dist2 := (leftEye.Col - rightEye.Col) * (leftEye.Col - rightEye.Col)
	dist := math.Sqrt(float64(dist1 + dist2))

	row := float64(leftEye.Row+rightEye.Row)/2.0 + 0.25*dist
	col := float64(leftEye.Col+rightEye.Col)/2.0 + 0.15*dist
	scale := 3.0 * dist

	flploc = &Puploc{
		Row:      int(row),
		Col:      int(col),
		Scale:    float32(scale),
		Perturbs: perturb,
	}

	if flipV {
		return plc.RunDetector(*flploc, img, 0.0, true)
	}
	return plc.RunDetector(*flploc, img, 0.0, false)
}

// ReadCascadeDir reads the facial landmark points cascade files from the provided directory.
func (plc *PuplocCascade) ReadCascadeDir(path string) (map[string][]*FlpCascade, error) {
	cascades, err := ioutil.ReadDir(path)
	if len(cascades) == 0 {
		return nil, errors.New("the provided directory is empty")
	}
	flpcs := make(map[string][]*FlpCascade)

	if err != nil {
		return nil, err
	}
	for _, cascade := range cascades {
		cf, err := filepath.Abs(path + "/" + cascade.Name())
		if err != nil {
			return nil, err
		}
		flpc, err := plc.UnpackFlp(cf)
		flpcs[cascade.Name()] = append(flpcs[cascade.Name()], &FlpCascade{flpc, err})
	}
	return flpcs, err
}
