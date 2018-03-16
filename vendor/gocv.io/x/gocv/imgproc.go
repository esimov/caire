package gocv

/*
#include <stdlib.h>
#include "imgproc.h"
*/
import "C"
import (
	"image"
	"image/color"
	"reflect"
	"unsafe"
)

// ArcLength calculates a contour perimeter or a curve length.
//
// For further details, please see:
//
// https://docs.opencv.org/master/d3/dc0/group__imgproc__shape.html#ga8d26483c636be6b35c3ec6335798a47c
//
func ArcLength(curve []image.Point, isClosed bool) float64 {
	cPoints := toCPoints(curve)
	arcLength := C.ArcLength(cPoints, C.bool(isClosed))
	return float64(arcLength)
}

// ApproxPolyDP approximates a polygonal curve(s) with the specified precision.
//
// For further details, please see:
//
// https://docs.opencv.org/master/d3/dc0/group__imgproc__shape.html#ga0012a5fdaea70b8a9970165d98722b4c
//
func ApproxPolyDP(curve []image.Point, epsilon float64, closed bool) (approxCurve []image.Point) {
	cCurve := toCPoints(curve)

	cApproxCurve := C.ApproxPolyDP(cCurve, C.double(epsilon), C.bool(closed))
	defer C.Points_Close(cApproxCurve)

	cApproxCurvePoints := (*[1 << 30]C.Point)(unsafe.Pointer(cApproxCurve.points))[:cApproxCurve.length:cApproxCurve.length]

	approxCurve = make([]image.Point, cApproxCurve.length)
	for i, cPoint := range cApproxCurvePoints {
		approxCurve[i] = image.Pt(int(cPoint.x), int(cPoint.y))
	}
	return approxCurve
}

// ConvexHull finds the convex hull of a point set.
//
// For further details, please see:
// https://docs.opencv.org/master/d3/dc0/group__imgproc__shape.html#ga014b28e56cb8854c0de4a211cb2be656
//
func ConvexHull(points []image.Point, hull Mat, clockwise bool, returnPoints bool) {
	cPoints := toCPoints(points)
	C.ConvexHull(cPoints, hull.p, C.bool(clockwise), C.bool(returnPoints))
}

// ConvexityDefects finds the convexity defects of a contour.
//
// For further details, please see:
// https://docs.opencv.org/master/d3/dc0/group__imgproc__shape.html#gada4437098113fd8683c932e0567f47ba
//
func ConvexityDefects(contour []image.Point, hull Mat, result Mat) {
	cPoints := toCPoints(contour)
	C.ConvexityDefects(cPoints, hull.p, result.p)
}

// CvtColor converts an image from one color space to another.
// It converts the src Mat image to the dst Mat using the
// code param containing the desired ColorConversionCode color space.
//
// For further details, please see:
// http://docs.opencv.org/master/d7/d1b/group__imgproc__misc.html#ga4e0972be5de079fed4e3a10e24ef5ef0
//
func CvtColor(src Mat, dst Mat, code ColorConversionCode) {
	C.CvtColor(src.p, dst.p, C.int(code))
}

// BilateralFilter applies a bilateral filter to an image.
//
// Bilateral filtering is described here:
// http://www.dai.ed.ac.uk/CVonline/LOCAL_COPIES/MANDUCHI1/Bilateral_Filtering.html
//
// BilateralFilter can reduce unwanted noise very well while keeping edges
// fairly sharp. However, it is very slow compared to most filters.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#ga9d7064d478c95d60003cf839430737ed
//
func BilateralFilter(src Mat, dst Mat, diameter int, sigmaColor float64, sigmaSpace float64) {
	C.BilateralFilter(src.p, dst.p, C.int(diameter), C.double(sigmaColor), C.double(sigmaSpace))
}

// Blur blurs an image Mat using a normalized box filter.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#ga8c45db9afe636703801b0b2e440fce37
//
func Blur(src Mat, dst Mat, ksize image.Point) {
	pSize := C.struct_Size{
		width:  C.int(ksize.X),
		height: C.int(ksize.Y),
	}

	C.Blur(src.p, dst.p, pSize)
}

// Dilate dilates an image by using a specific structuring element.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#ga4ff0f3318642c4f469d0e11f242f3b6c
//
func Dilate(src Mat, dst Mat, kernel Mat) {
	C.Dilate(src.p, dst.p, kernel.p)
}

// Erode erodes an image by using a specific structuring element.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gaeb1e0c1033e3f6b891a25d0511362aeb
//
func Erode(src Mat, dst Mat, kernel Mat) {
	C.Erode(src.p, dst.p, kernel.p)
}

// RetrievalMode is the mode of the contour retrieval algorithm.
type RetrievalMode int

const (
	// RetrievalExternal retrieves only the extreme outer contours.
	// It sets `hierarchy[i][2]=hierarchy[i][3]=-1` for all the contours.
	RetrievalExternal RetrievalMode = 0

	// RetrievalList retrieves all of the contours without establishing
	// any hierarchical relationships.
	RetrievalList = 1

	// RetrievalCComp retrieves all of the contours and organizes them into
	// a two-level hierarchy. At the top level, there are external boundaries
	// of the components. At the second level, there are boundaries of the holes.
	// If there is another contour inside a hole of a connected component, it
	// is still put at the top level.
	RetrievalCComp = 2

	// RetrievalTree retrieves all of the contours and reconstructs a full
	// hierarchy of nested contours.
	RetrievalTree = 3

	// RetrievalFloodfill lacks a description in the original header.
	RetrievalFloodfill = 4
)

// ContourApproximationMode is the mode of the contour approximation algorithm.
type ContourApproximationMode int

const (
	// ChainApproxNone stores absolutely all the contour points. That is,
	// any 2 subsequent points (x1,y1) and (x2,y2) of the contour will be
	// either horizontal, vertical or diagonal neighbors, that is,
	// max(abs(x1-x2),abs(y2-y1))==1.
	ChainApproxNone ContourApproximationMode = 1

	// ChainApproxSimple compresses horizontal, vertical, and diagonal segments
	// and leaves only their end points.
	// For example, an up-right rectangular contour is encoded with 4 points.
	ChainApproxSimple = 2

	// ChainApproxTC89L1 applies one of the flavors of the Teh-Chin chain
	// approximation algorithms.
	ChainApproxTC89L1 = 3

	// ChainApproxTC89KCOS applies one of the flavors of the Teh-Chin chain
	// approximation algorithms.
	ChainApproxTC89KCOS = 4
)

// BoundingRect calculates the up-right bounding rectangle of a point set.
//
// For further details, please see:
// https://docs.opencv.org/3.3.0/d3/dc0/group__imgproc__shape.html#gacb413ddce8e48ff3ca61ed7cf626a366
//
func BoundingRect(contour []image.Point) image.Rectangle {
	cContour := toCPoints(contour)
	r := C.BoundingRect(cContour)
	rect := image.Rect(int(r.x), int(r.y), int(r.x+r.width), int(r.y+r.height))
	return rect
}

// ContourArea calculates a contour area.
//
// For further details, please see:
// https://docs.opencv.org/3.3.0/d3/dc0/group__imgproc__shape.html#ga2c759ed9f497d4a618048a2f56dc97f1
//
func ContourArea(contour []image.Point) float64 {
	cContour := toCPoints(contour)
	result := C.ContourArea(cContour)
	return float64(result)
}

// FindContours finds contours in a binary image.
//
// For further details, please see:
// https://docs.opencv.org/3.3.0/d3/dc0/group__imgproc__shape.html#ga17ed9f5d79ae97bd4c7cf18403e1689a
//
func FindContours(src Mat, mode RetrievalMode, method ContourApproximationMode) [][]image.Point {
	ret := C.FindContours(src.p, C.int(mode), C.int(method))
	defer C.Contours_Close(ret)

	cArray := ret.contours
	cLength := int(ret.length)
	cHdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cArray)),
		Len:  cLength,
		Cap:  cLength,
	}
	sContours := *(*[]C.Points)(unsafe.Pointer(&cHdr))

	contours := make([][]image.Point, cLength)
	for i, pts := range sContours {
		pArray := pts.points
		pLength := int(pts.length)
		pHdr := reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(pArray)),
			Len:  pLength,
			Cap:  pLength,
		}
		sPoints := *(*[]C.Point)(unsafe.Pointer(&pHdr))

		points := make([]image.Point, pLength)
		for j, pt := range sPoints {
			points[j] = image.Pt(int(pt.x), int(pt.y))
		}
		contours[i] = points
	}

	return contours
}

// TemplateMatchMode is the type of the template matching operation.
type TemplateMatchMode int

const (
	// TmSqdiff maps to TM_SQDIFF
	TmSqdiff TemplateMatchMode = 0
	// TmSqdiffNormed maps to TM_SQDIFF_NORMED
	TmSqdiffNormed = 1
	// TmCcorr maps to TM_CCORR
	TmCcorr = 2
	// TmCcorrNormed maps to TM_CCORR_NORMED
	TmCcorrNormed = 3
	// TmCcoeff maps to TM_CCOEFF
	TmCcoeff = 4
	// TmCcoeffNormed maps to TM_CCOEFF_NORMED
	TmCcoeffNormed = 5
)

// MatchTemplate compares a template against overlapped image regions.
//
// For further details, please see:
// https://docs.opencv.org/master/df/dfb/group__imgproc__object.html#ga586ebfb0a7fb604b35a23d85391329be
//
func MatchTemplate(image Mat, templ Mat, result Mat, method TemplateMatchMode, mask Mat) {
	C.MatchTemplate(image.p, templ.p, result.p, C.int(method), mask.p)
}

// Moments calculates all of the moments up to the third order of a polygon
// or rasterized shape.
//
// For further details, please see:
// https://docs.opencv.org/master/d3/dc0/group__imgproc__shape.html#ga556a180f43cab22649c23ada36a8a139
//
func Moments(src Mat, binaryImage bool) map[string]float64 {
	r := C.Moments(src.p, C.bool(binaryImage))

	result := make(map[string]float64)
	result["m00"] = float64(r.m00)
	result["m10"] = float64(r.m10)
	result["m01"] = float64(r.m01)
	result["m20"] = float64(r.m20)
	result["m11"] = float64(r.m11)
	result["m02"] = float64(r.m02)
	result["m30"] = float64(r.m30)
	result["m21"] = float64(r.m21)
	result["m12"] = float64(r.m12)
	result["m03"] = float64(r.m03)
	result["mu20"] = float64(r.mu20)
	result["mu11"] = float64(r.mu11)
	result["mu02"] = float64(r.mu02)
	result["mu30"] = float64(r.mu30)
	result["mu21"] = float64(r.mu21)
	result["mu12"] = float64(r.mu12)
	result["mu03"] = float64(r.mu03)
	result["nu20"] = float64(r.nu20)
	result["nu11"] = float64(r.nu11)
	result["nu02"] = float64(r.nu02)
	result["nu30"] = float64(r.nu30)
	result["nu21"] = float64(r.nu21)
	result["nu12"] = float64(r.nu12)
	result["nu03"] = float64(r.nu03)

	return result
}

// PyrDown blurs an image and downsamples it.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gaf9bba239dfca11654cb7f50f889fc2ff
//
func PyrDown(src Mat, dst Mat, ksize image.Point, borderType BorderType) {
	pSize := C.struct_Size{
		height: C.int(ksize.X),
		width:  C.int(ksize.Y),
	}
	C.PyrDown(src.p, dst.p, pSize, C.int(borderType))
}

// PyrUp upsamples an image and then blurs it.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gada75b59bdaaca411ed6fee10085eb784
//
func PyrUp(src Mat, dst Mat, ksize image.Point, borderType BorderType) {
	pSize := C.struct_Size{
		height: C.int(ksize.X),
		width:  C.int(ksize.Y),
	}
	C.PyrUp(src.p, dst.p, pSize, C.int(borderType))
}

// MorphologyEx performs advanced morphological transformations.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#ga67493776e3ad1a3df63883829375201f
//
func MorphologyEx(src Mat, dst Mat, op MorphType, kernel Mat) {
	C.MorphologyEx(src.p, dst.p, C.int(op), kernel.p)
}

// MorphShape is the shape of the structuring element used for Morphing operations.
type MorphShape int

const (
	// MorphRect is the rectangular morph shape.
	MorphRect MorphShape = 0

	// MorphCross is the cross morph shape.
	MorphCross = 1

	// MorphEllipse is the ellipse morph shape.
	MorphEllipse = 2
)

// GetStructuringElement returns a structuring element of the specified size
// and shape for morphological operations.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gac342a1bb6eabf6f55c803b09268e36dc
//
func GetStructuringElement(shape MorphShape, ksize image.Point) Mat {
	sz := C.struct_Size{
		width:  C.int(ksize.X),
		height: C.int(ksize.Y),
	}

	return Mat{p: C.GetStructuringElement(C.int(shape), sz)}
}

// MorphType type of morphological operation.
type MorphType int

const (
	// MorphErode operation
	MorphErode MorphType = 0

	// MorphDilate operation
	MorphDilate = 1

	// MorphOpen operation
	MorphOpen = 2

	// MorphClose operation
	MorphClose = 3

	// MorphGradient operation
	MorphGradient = 4

	// MorphTophat operation
	MorphTophat = 5

	// MorphBlackhat operation
	MorphBlackhat = 6

	// MorphHitmiss operation
	MorphHitmiss = 7
)

// BorderType type of border.
type BorderType int

const (
	// BorderConstant border type
	BorderConstant BorderType = 0

	// BorderReplicate border type
	BorderReplicate = 1

	// BorderReflect border type
	BorderReflect = 2

	// BorderWrap border type
	BorderWrap = 3

	// BorderReflect101 border type
	BorderReflect101 = 4

	// BorderTransparent border type
	BorderTransparent = 5

	// BorderDefault border type
	BorderDefault = BorderReflect101
)

// GaussianBlur blurs an image Mat using a Gaussian filter.
// The function convolves the src Mat image into the dst Mat using
// the specified Gaussian kernel params.
//
// For further details, please see:
// http://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gaabe8c836e97159a9193fb0b11ac52cf1
//
func GaussianBlur(src Mat, dst Mat, ksize image.Point, sigmaX float64,
	sigmaY float64, borderType BorderType) {
	pSize := C.struct_Size{
		width:  C.int(ksize.X),
		height: C.int(ksize.Y),
	}

	C.GaussianBlur(src.p, dst.p, pSize, C.double(sigmaX), C.double(sigmaY), C.int(borderType))
}

// Laplacian calculates the Laplacian of an image.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gad78703e4c8fe703d479c1860d76429e6
//
func Laplacian(src Mat, dst Mat, dDepth int, size int, scale float64,
	delta float64, borderType BorderType) {
	C.Laplacian(src.p, dst.p, C.int(dDepth), C.int(size), C.double(scale), C.double(delta), C.int(borderType))
}

// Scharr calculates the first x- or y- image derivative using Scharr operator.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#gaa13106761eedf14798f37aa2d60404c9
//
func Scharr(src Mat, dst Mat, dDepth int, dx int, dy int, scale float64,
	delta float64, borderType BorderType) {
	C.Scharr(src.p, dst.p, C.int(dDepth), C.int(dx), C.int(dy), C.double(scale), C.double(delta), C.int(borderType))
}

// MedianBlur blurs an image using the median filter.
//
// For further details, please see:
// https://docs.opencv.org/master/d4/d86/group__imgproc__filter.html#ga564869aa33e58769b4469101aac458f9
//
func MedianBlur(src Mat, dst Mat, ksize int) {
	C.MedianBlur(src.p, dst.p, C.int(ksize))
}

// Canny finds edges in an image using the Canny algorithm.
// The function finds edges in the input image image and marks
// them in the output map edges using the Canny algorithm.
// The smallest value between threshold1 and threshold2 is used
// for edge linking. The largest value is used to
// find initial segments of strong edges.
// See http://en.wikipedia.org/wiki/Canny_edge_detector
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d1a/group__imgproc__feature.html#ga04723e007ed888ddf11d9ba04e2232de
//
func Canny(src Mat, edges Mat, t1 float32, t2 float32) {
	C.Canny(src.p, edges.p, C.double(t1), C.double(t2))
}

// CornerSubPix Refines the corner locations. The function iterates to find
// the sub-pixel accurate location of corners or radial saddle points.
//
// For further details, please see:
// https://docs.opencv.org/master/dd/d1a/group__imgproc__feature.html#ga354e0d7c86d0d9da75de9b9701a9a87e
//
func CornerSubPix(img Mat, corners Mat, winSize image.Point, zeroZone image.Point, criteria TermCriteria) {
	winSz := C.struct_Size{
		width:  C.int(winSize.X),
		height: C.int(winSize.Y),
	}

	zeroSz := C.struct_Size{
		width:  C.int(zeroZone.X),
		height: C.int(zeroZone.Y),
	}

	C.CornerSubPix(img.p, corners.p, winSz, zeroSz, criteria.p)
	return
}

// GoodFeaturesToTrack determines strong corners on an image. The function
// finds the most prominent corners in the image or in the specified image region.
//
// For further details, please see:
// https://docs.opencv.org/master/dd/d1a/group__imgproc__feature.html#ga1d6bb77486c8f92d79c8793ad995d541
//
func GoodFeaturesToTrack(img Mat, corners Mat, maxCorners int, quality float64, minDist float64) {
	C.GoodFeaturesToTrack(img.p, corners.p, C.int(maxCorners), C.double(quality), C.double(minDist))
}

// HoughCircles finds circles in a grayscale image using the Hough transform.
// The only "method" currently supported is HOUGH_GRADIENT = 3.
//
// For further details, please see:
// https://docs.opencv.org/master/dd/d1a/group__imgproc__feature.html#ga47849c3be0d0406ad3ca45db65a25d2d
//
func HoughCircles(src Mat, circles Mat, method int, dp float64, minDist float64) {
	C.HoughCircles(src.p, circles.p, C.int(method), C.double(dp), C.double(minDist))
}

// HoughLines implements the standard or standard multi-scale Hough transform
// algorithm for line detection. For a good explanation of Hough transform, see:
// http://homepages.inf.ed.ac.uk/rbf/HIPR2/hough.htm
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d1a/group__imgproc__feature.html#ga46b4e588934f6c8dfd509cc6e0e4545a
//
func HoughLines(src Mat, lines Mat, rho float32, theta float32, threshold int) {
	C.HoughLines(src.p, lines.p, C.double(rho), C.double(theta), C.int(threshold))
}

// HoughLinesP implements the probabilistic Hough transform
// algorithm for line detection. For a good explanation of Hough transform, see:
// http://homepages.inf.ed.ac.uk/rbf/HIPR2/hough.htm
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d1a/group__imgproc__feature.html#ga8618180a5948286384e3b7ca02f6feeb
//
func HoughLinesP(src Mat, lines Mat, rho float32, theta float32, threshold int) {
	C.HoughLinesP(src.p, lines.p, C.double(rho), C.double(theta), C.int(threshold))
}

// ThresholdType type of threshold operation.
type ThresholdType int

const (
	// ThresholdBinary threshold type
	ThresholdBinary ThresholdType = 0

	// ThresholdBinaryInv threshold type
	ThresholdBinaryInv = 1

	// ThresholdTrunc threshold type
	ThresholdTrunc = 2

	// ThresholdToZero threshold type
	ThresholdToZero = 3

	// ThresholdToZeroInv threshold type
	ThresholdToZeroInv = 4

	// ThresholdMask threshold type
	ThresholdMask = 7

	// ThresholdOtsu threshold type
	ThresholdOtsu = 8

	// ThresholdTriangle threshold type
	ThresholdTriangle = 16
)

// Threshold applies a fixed-level threshold to each array element.
//
// For further details, please see:
// https://docs.opencv.org/3.3.0/d7/d1b/group__imgproc__misc.html#gae8a4a146d1ca78c626a53577199e9c57
//
func Threshold(src Mat, dst Mat, thresh float32, maxvalue float32, typ ThresholdType) {
	C.Threshold(src.p, dst.p, C.double(thresh), C.double(maxvalue), C.int(typ))
}

// AdaptiveThresholdType type of adaptive threshold operation.
type AdaptiveThresholdType int

const (
	// AdaptiveThresholdMean threshold type
	AdaptiveThresholdMean AdaptiveThresholdType = 0

	// AdaptiveThresholdGaussian threshold type
	AdaptiveThresholdGaussian = 1
)

// AdaptiveThreshold applies a fixed-level threshold to each array element.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/d1b/group__imgproc__misc.html#ga72b913f352e4a1b1b397736707afcde3
//
func AdaptiveThreshold(src Mat, dst Mat, maxValue float32, adaptiveTyp AdaptiveThresholdType, typ ThresholdType, blockSize int, c float32) {
	C.AdaptiveThreshold(src.p, dst.p, C.double(maxValue), C.int(adaptiveTyp), C.int(typ), C.int(blockSize), C.double(c))
}

// ArrowedLine draws a arrow segment pointing from the first point
// to the second one.
//
// For further details, please see:
// https://docs.opencv.org/master/d6/d6e/group__imgproc__draw.html#ga0a165a3ca093fd488ac709fdf10c05b2
//
func ArrowedLine(img Mat, pt1 image.Point, pt2 image.Point, c color.RGBA, thickness int) {
	sp1 := C.struct_Point{
		x: C.int(pt1.X),
		y: C.int(pt1.Y),
	}

	sp2 := C.struct_Point{
		x: C.int(pt2.X),
		y: C.int(pt2.Y),
	}

	sColor := C.struct_Scalar{
		val1: C.double(c.B),
		val2: C.double(c.G),
		val3: C.double(c.R),
		val4: C.double(c.A),
	}

	C.ArrowedLine(img.p, sp1, sp2, sColor, C.int(thickness))
}

// Circle draws a circle.
//
// For further details, please see:
// https://docs.opencv.org/master/d6/d6e/group__imgproc__draw.html#gaf10604b069374903dbd0f0488cb43670
//
func Circle(img Mat, center image.Point, radius int, c color.RGBA, thickness int) {
	pc := C.struct_Point{
		x: C.int(center.X),
		y: C.int(center.Y),
	}

	sColor := C.struct_Scalar{
		val1: C.double(c.B),
		val2: C.double(c.G),
		val3: C.double(c.R),
		val4: C.double(c.A),
	}

	C.Circle(img.p, pc, C.int(radius), sColor, C.int(thickness))
}

// Line draws a line segment connecting two points.
//
// For further details, please see:
// https://docs.opencv.org/master/d6/d6e/group__imgproc__draw.html#ga7078a9fae8c7e7d13d24dac2520ae4a2
//
func Line(img Mat, pt1 image.Point, pt2 image.Point, c color.RGBA, thickness int) {
	sp1 := C.struct_Point{
		x: C.int(pt1.X),
		y: C.int(pt1.Y),
	}

	sp2 := C.struct_Point{
		x: C.int(pt2.X),
		y: C.int(pt2.Y),
	}

	sColor := C.struct_Scalar{
		val1: C.double(c.B),
		val2: C.double(c.G),
		val3: C.double(c.R),
		val4: C.double(c.A),
	}

	C.Line(img.p, sp1, sp2, sColor, C.int(thickness))
}

// Rectangle draws a simple, thick, or filled up-right rectangle.
// It renders a rectangle with the desired characteristics to the target Mat image.
//
// For further details, please see:
// http://docs.opencv.org/master/d6/d6e/group__imgproc__draw.html#ga346ac30b5c74e9b5137576c9ee9e0e8c
//
func Rectangle(img Mat, r image.Rectangle, c color.RGBA, thickness int) {
	cRect := C.struct_Rect{
		x:      C.int(r.Min.X),
		y:      C.int(r.Min.Y),
		width:  C.int(r.Size().X),
		height: C.int(r.Size().Y),
	}

	sColor := C.struct_Scalar{
		val1: C.double(c.B),
		val2: C.double(c.G),
		val3: C.double(c.R),
		val4: C.double(c.A),
	}

	C.Rectangle(img.p, cRect, sColor, C.int(thickness))
}

// HersheyFont are the font libraries included in OpenCV.
// Only a subset of the available Hershey fonts are supported by OpenCV.
//
// For more information, see:
// http://sources.isc.org/utils/misc/hershey-font.txt
//
type HersheyFont int

const (
	// FontHersheySimplex is normal size sans-serif font.
	FontHersheySimplex HersheyFont = 0
	// FontHersheyPlain issmall size sans-serif font.
	FontHersheyPlain = 1
	// FontHersheyDuplex normal size sans-serif font
	// (more complex than FontHersheySIMPLEX).
	FontHersheyDuplex = 2
	// FontHersheyComplex i a normal size serif font.
	FontHersheyComplex = 3
	// FontHersheyTriplex is a normal size serif font
	// (more complex than FontHersheyCOMPLEX).
	FontHersheyTriplex = 4
	// FontHersheyComplexSmall is a smaller version of FontHersheyCOMPLEX.
	FontHersheyComplexSmall = 5
	// FontHersheyScriptSimplex is a hand-writing style font.
	FontHersheyScriptSimplex = 6
	// FontHersheyScriptComplex is a more complex variant of FontHersheyScriptSimplex.
	FontHersheyScriptComplex = 7
	// FontItalic is the flag for italic font.
	FontItalic = 16
)

// GetTextSize calculates the width and height of a text string.
// It returns an image.Point with the size required to draw text using
// a specific font face, scale, and thickness.
//
// For further details, please see:
// http://docs.opencv.org/master/d6/d6e/group__imgproc__draw.html#ga3d2abfcb995fd2db908c8288199dba82
//
func GetTextSize(text string, fontFace HersheyFont, fontScale float64, thickness int) image.Point {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	sz := C.GetTextSize(cText, C.int(fontFace), C.double(fontScale), C.int(thickness))
	return image.Pt(int(sz.width), int(sz.height))
}

// PutText draws a text string.
// It renders the specified text string into the img Mat at the location
// passed in the "org" param, using the desired font face, font scale,
// color, and line thinkness.
//
// For further details, please see:
// http://docs.opencv.org/master/d6/d6e/group__imgproc__draw.html#ga5126f47f883d730f633d74f07456c576
//
func PutText(img Mat, text string, org image.Point, fontFace HersheyFont, fontScale float64, c color.RGBA, thickness int) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	pOrg := C.struct_Point{
		x: C.int(org.X),
		y: C.int(org.Y),
	}

	sColor := C.struct_Scalar{
		val1: C.double(c.B),
		val2: C.double(c.G),
		val3: C.double(c.R),
		val4: C.double(c.A),
	}

	C.PutText(img.p, cText, pOrg, C.int(fontFace), C.double(fontScale), sColor, C.int(thickness))
	return
}

// InterpolationFlags are bit flags that control the interpolation algorithm
// that is used.
type InterpolationFlags int

const (
	// InterpolationNearestNeighbor is nearest neighbor. (fast but low quality)
	InterpolationNearestNeighbor InterpolationFlags = 0

	// InterpolationLinear is bilinear interpolation.
	InterpolationLinear = 1

	// InterpolationCubic is bicube interpolation.
	InterpolationCubic = 2

	// InterpolationArea uses pixel area relation. It is preferred for image
	// decimation as it gives moire-free results.
	InterpolationArea = 3

	// InterpolationLanczos4 is Lanczos interpolation over 8x8 neighborhood.
	InterpolationLanczos4 = 4

	// InterpolationDefault is an alias for InterpolationLinear.
	InterpolationDefault = InterpolationLinear

	// InterpolationMax indicates use maximum interpolation.
	InterpolationMax = 7
)

// Resize resizes an image.
// It resizes the image src down to or up to the specified size, storing the
// result in dst. Note that src and dst may be the same image. If you wish to
// scale by factor, an empty sz may be passed and non-zero fx and fy. Likewise,
// if you wish to scale to an explicit size, a non-empty sz may be passed with
// zero for both fx and fy.
//
// For further details, please see:
// https://docs.opencv.org/master/da/d54/group__imgproc__transform.html#ga47a974309e9102f5f08231edc7e7529d
func Resize(src, dst Mat, sz image.Point, fx, fy float64, interp InterpolationFlags) {
	pSize := C.struct_Size{
		width:  C.int(sz.X),
		height: C.int(sz.Y),
	}

	C.Resize(src.p, dst.p, pSize, C.double(fx), C.double(fy), C.int(interp))
	return
}

// GetRotationMatrix2D calculates an affine matrix of 2D rotation.
//
// For further details, please see:
// https://docs.opencv.org/master/da/d54/group__imgproc__transform.html#gafbbc470ce83812914a70abfb604f4326
func GetRotationMatrix2D(center image.Point, angle, scale float64) Mat {
	pc := C.struct_Point{
		x: C.int(center.X),
		y: C.int(center.Y),
	}
	return Mat{p: C.GetRotationMatrix2D(pc, C.double(angle), C.double(scale))}
}

// WarpAffine applies an affine transformation to an image. For more parameters please check WarpAffineWithParams
//
// For further details, please see:
// https://docs.opencv.org/master/da/d54/group__imgproc__transform.html#ga0203d9ee5fcd28d40dbc4a1ea4451983
func WarpAffine(src, dst, m Mat, sz image.Point) {
	pSize := C.struct_Size{
		width:  C.int(sz.X),
		height: C.int(sz.Y),
	}

	C.WarpAffine(src.p, dst.p, m.p, pSize)
}

// WarpAffineWithParams applies an affine transformation to an image.
//
// For further details, please see:
// https://docs.opencv.org/master/da/d54/group__imgproc__transform.html#ga0203d9ee5fcd28d40dbc4a1ea4451983
func WarpAffineWithParams(src, dst, m Mat, sz image.Point, flags InterpolationFlags, borderType BorderType, borderValue color.RGBA) {
	pSize := C.struct_Size{
		width:  C.int(sz.X),
		height: C.int(sz.Y),
	}
	bv := C.struct_Scalar{
		val1: C.double(borderValue.B),
		val2: C.double(borderValue.G),
		val3: C.double(borderValue.R),
		val4: C.double(borderValue.A),
	}
	C.WarpAffineWithParams(src.p, dst.p, m.p, pSize, C.int(flags), C.int(borderType), bv)
}

// WarpPerspective applies a perspective transformation to an image.
//
// For further details, please see:
// https://docs.opencv.org/master/da/d54/group__imgproc__transform.html#gaf73673a7e8e18ec6963e3774e6a94b87
func WarpPerspective(src, dst, m Mat, sz image.Point) {
	pSize := C.struct_Size{
		width:  C.int(sz.X),
		height: C.int(sz.Y),
	}

	C.WarpPerspective(src.p, dst.p, m.p, pSize)
}

// ColormapTypes are the 12 GNU Octave/MATLAB equivalent colormaps.
//
// For further details, please see:
// https://docs.opencv.org/master/d3/d50/group__imgproc__colormap.html
type ColormapTypes int

// List of the available color maps
//
// For further details, please see:
// https://docs.opencv.org/master/d3/d50/group__imgproc__colormap.html#ga9a805d8262bcbe273f16be9ea2055a65
const (
	ColormapAutumn  ColormapTypes = 0
	ColormapBone                  = 1
	ColormapJet                   = 2
	ColormapWinter                = 3
	ColormapRainbow               = 4
	ColormapOcean                 = 5
	ColormapSummer                = 6
	ColormapSpring                = 7
	ColormapCool                  = 8
	ColormapHsv                   = 9
	ColormapPink                  = 10
	ColormapHot                   = 11
	ColormapParula                = 12
)

// ApplyColorMap applies a GNU Octave/MATLAB equivalent colormap on a given image.
//
// For further details, please see:
// https://docs.opencv.org/master/d3/d50/group__imgproc__colormap.html#gadf478a5e5ff49d8aa24e726ea6f65d15
func ApplyColorMap(src, dst Mat, colormapType ColormapTypes) {
	C.ApplyColorMap(src.p, dst.p, C.int(colormapType))
}

// ApplyCustomColorMap applies a custom defined colormap on a given image.
//
// For further details, please see:
// https://docs.opencv.org/master/d3/d50/group__imgproc__colormap.html#gacb22288ddccc55f9bd9e6d492b409cae
func ApplyCustomColorMap(src, dst, customColormap Mat) {
	C.ApplyCustomColorMap(src.p, dst.p, customColormap.p)
}

// GetPerspectiveTransform returns 3x3 perspective transformation for the
// corresponding 4 point pairs.
//
// For further details, please see:
// https://docs.opencv.org/master/da/d54/group__imgproc__transform.html#ga8c1ae0e3589a9d77fffc962c49b22043
func GetPerspectiveTransform(src, dst []image.Point) Mat {
	srcPoints := toCPoints(src)
	dstPoints := toCPoints(dst)
	return Mat{p: C.GetPerspectiveTransform(srcPoints, dstPoints)}
}

// DrawContours draws contours outlines or filled contours.
//
// For further details, please see:
// https://docs.opencv.org/3.3.1/d6/d6e/group__imgproc__draw.html#ga746c0625f1781f1ffc9056259103edbc
func DrawContours(src Mat, contours [][]image.Point, contourIdx int, c color.RGBA, thickness int) {
	cntrs := make([]C.struct_Points, len(contours))

	for i, contour := range contours {
		p := (*C.struct_Point)(C.malloc(C.size_t(C.sizeof_struct_Point * len(contour))))
		defer C.free(unsafe.Pointer(p))

		pa := (*[1 << 30]C.struct_Point)(unsafe.Pointer(p))

		for j, point := range contour {
			(*pa)[j] = C.struct_Point{
				x: C.int(point.X),
				y: C.int(point.Y),
			}
		}

		cntrs[i] = C.struct_Points{
			points: (*C.Point)(p),
			length: C.int(len(contour)),
		}
	}

	cContours := C.struct_Contours{
		contours: (*C.struct_Points)(&cntrs[0]),
		length:   C.int(len(contours)),
	}

	sColor := C.struct_Scalar{
		val1: C.double(c.B),
		val2: C.double(c.G),
		val3: C.double(c.R),
		val4: C.double(c.A),
	}

	C.DrawContours(src.p, cContours, C.int(contourIdx), sColor, C.int(thickness))
}
