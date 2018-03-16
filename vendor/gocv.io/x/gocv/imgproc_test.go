package gocv

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"
)

func TestApproxPolyDP(t *testing.T) {
	img := NewMatWithSize(100, 200, MatTypeCV8UC1)
	defer img.Close()

	white := color.RGBA{255, 255, 255, 255}
	// Draw triangle
	Line(img, image.Pt(25, 25), image.Pt(25, 75), white, 1)
	Line(img, image.Pt(25, 75), image.Pt(75, 50), white, 1)
	Line(img, image.Pt(75, 50), image.Pt(25, 25), white, 1)
	// Draw rectangle
	Rectangle(img, image.Rect(125, 25, 175, 75), white, 1)

	contours := FindContours(img, RetrievalExternal, ChainApproxSimple)

	trianglePerimeter := ArcLength(contours[0], true)
	triangleContour := ApproxPolyDP(contours[0], 0.04*trianglePerimeter, true)
	expectedTriangleContour := []image.Point{image.Pt(25, 25), image.Pt(25, 75), image.Pt(75, 50)}
	if !reflect.DeepEqual(triangleContour, expectedTriangleContour) {
		t.Errorf("Failed to approximate triangle.\nActual:%v\nExpect:%v", triangleContour, expectedTriangleContour)
	}

	rectPerimeter := ArcLength(contours[1], true)
	rectContour := ApproxPolyDP(contours[1], 0.04*rectPerimeter, true)
	expectedRectContour := []image.Point{image.Pt(125, 24), image.Pt(124, 75), image.Pt(175, 76), image.Pt(176, 25)}
	if !reflect.DeepEqual(rectContour, expectedRectContour) {
		t.Errorf("Failed to approximate rectangle.\nActual:%v\nExpect:%v", rectContour, expectedRectContour)
	}
}

func TestConvexity(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in FindContours test")
	}
	defer img.Close()

	res := FindContours(img, RetrievalExternal, ChainApproxSimple)
	if len(res) < 1 {
		t.Error("Invalid FindContours test")
	}

	area := ContourArea(res[0])
	if area != 127280.0 {
		t.Errorf("Invalid ContourArea test: %f", area)
	}

	hull := NewMat()
	defer hull.Close()

	ConvexHull(res[0], hull, true, false)
	if hull.Empty() {
		t.Error("Invalid ConvexHull test")
	}

	defects := NewMat()
	defer defects.Close()

	ConvexityDefects(res[0], hull, defects)
	if defects.Empty() {
		t.Error("Invalid ConvexityDefects test")
	}
}

func TestCvtColor(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in CvtColor test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	CvtColor(img, dest, ColorBGRAToGray)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid convert in CvtColor test")
	}
}

func TestBilateralFilter(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in BilateralFilter test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	BilateralFilter(img, dest, 1, 2.0, 3.0)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid BilateralFilter test")
	}
}

func TestBlur(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in GaussianBlur test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	Blur(img, dest, image.Pt(3, 3))
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Blur test")
	}
}

func TestDilate(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in Dilate test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	kernel := GetStructuringElement(MorphRect, image.Pt(1, 1))

	Dilate(img, dest, kernel)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Dilate test")
	}
}

func TestMatchTemplate(t *testing.T) {
	imgScene := IMRead("images/face.jpg", IMReadGrayScale)
	if imgScene.Empty() {
		t.Error("Invalid read of face.jpg in MatchTemplate test")
	}
	defer imgScene.Close()

	imgTemplate := IMRead("images/toy.jpg", IMReadGrayScale)
	if imgTemplate.Empty() {
		t.Error("Invalid read of toy.jpg in MatchTemplate test")
	}
	defer imgTemplate.Close()

	result := NewMat()
	MatchTemplate(imgScene, imgTemplate, result, TmCcoeffNormed, NewMat())
	_, maxConfidence, _, _ := MinMaxLoc(result)
	if maxConfidence < 0.95 {
		t.Errorf("Max confidence of %f is too low. MatchTemplate could not find template in scene.", maxConfidence)
	}
}

func TestMoments(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in Moments test")
	}
	defer img.Close()

	result := Moments(img, true)
	if len(result) < 1 {
		t.Errorf("Invalid Moments test: %v", result)
	}
}

func TestPyrDown(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in PyrDown test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	PyrDown(img, dest, image.Point{X: dest.Cols(), Y: dest.Rows()}, BorderDefault)
	if dest.Empty() && math.Abs(float64(img.Cols()-2*dest.Cols())) < 2.0 && math.Abs(float64(img.Rows()-2*dest.Rows())) < 2.0 {
		t.Error("Invalid PyrDown test")
	}
}

func TestPyrUp(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in PyrUp test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	PyrUp(img, dest, image.Point{X: dest.Cols(), Y: dest.Rows()}, BorderDefault)
	if dest.Empty() && math.Abs(float64(2*img.Cols()-dest.Cols())) < 2.0 && math.Abs(float64(2*img.Rows()-dest.Rows())) < 2.0 {
		t.Error("Invalid PyrUp test")
	}
}

func TestFindContours(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in FindContours test")
	}
	defer img.Close()

	res := FindContours(img, RetrievalExternal, ChainApproxSimple)
	if len(res) < 1 {
		t.Error("Invalid FindContours test")
	}

	area := ContourArea(res[0])
	if area != 127280.0 {
		t.Errorf("Invalid ContourArea test: %f", area)
	}

	r := BoundingRect(res[0])
	if !r.Eq(image.Rect(0, 0, 400, 320)) {
		t.Errorf("Invalid BoundingRect test: %v", r)
	}

	length := ArcLength(res[0], true)
	if int(length) != 1436 {
		t.Errorf("Invalid ArcLength test: %f", length)
	}

	length = ArcLength(res[0], false)
	if int(length) != 1037 {
		t.Errorf("Invalid ArcLength test: %f", length)
	}
}

func TestErode(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in Erode test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	kernel := GetStructuringElement(MorphRect, image.Pt(1, 1))

	Erode(img, dest, kernel)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Erode test")
	}
}

func TestMorphologyEx(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in MorphologyEx test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	kernel := GetStructuringElement(MorphRect, image.Pt(1, 1))

	MorphologyEx(img, dest, MorphOpen, kernel)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid MorphologyEx test")
	}
}

func TestGaussianBlur(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in GaussianBlur test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	GaussianBlur(img, dest, image.Pt(23, 23), 30, 50, 4)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Blur test")
	}
}

func TestLaplacian(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in Laplacian test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	Laplacian(img, dest, MatTypeCV16S, 1, 1, 0, BorderDefault)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Laplacian test")
	}
}

func TestScharr(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in Scharr test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	Scharr(img, dest, MatTypeCV16S, 1, 0, 0, 0, BorderDefault)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Scharr test")
	}
}

func TestMedianBlur(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in MedianBlur test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	MedianBlur(img, dest, 1)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid MedianBlur test")
	}
}

func TestCanny(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in HoughLines test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	Canny(img, dest, 50, 150)
	if dest.Empty() {
		t.Error("Empty Canny test")
	}
	if img.Rows() != dest.Rows() {
		t.Error("Invalid Canny test rows")
	}
	if img.Cols() != dest.Cols() {
		t.Error("Invalid Canny test cols")
	}
}

func TestGoodFeaturesToTrackAndCornerSubPix(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in GoodFeaturesToTrack test")
	}
	defer img.Close()

	corners := NewMat()
	defer corners.Close()

	GoodFeaturesToTrack(img, corners, 500, 0.01, 10)
	if corners.Empty() {
		t.Error("Empty GoodFeaturesToTrack test")
	}
	if corners.Rows() != 205 {
		t.Errorf("Invalid GoodFeaturesToTrack test rows: %v", corners.Rows())
	}
	if corners.Cols() != 1 {
		t.Errorf("Invalid GoodFeaturesToTrack test cols: %v", corners.Cols())
	}

	tc := NewTermCriteria(Count|EPS, 20, 0.03)

	CornerSubPix(img, corners, image.Pt(10, 10), image.Pt(-1, -1), tc)
	if corners.Empty() {
		t.Error("Empty CornerSubPix test")
	}
	if corners.Rows() != 205 {
		t.Errorf("Invalid CornerSubPix test rows: %v", corners.Rows())
	}
	if corners.Cols() != 1 {
		t.Errorf("Invalid CornerSubPix test cols: %v", corners.Cols())
	}
}

func TestHoughCircles(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in HoughCircles test")
	}
	defer img.Close()

	circles := NewMat()
	defer circles.Close()

	HoughCircles(img, circles, 3, 5.0, 5.0)
	if circles.Empty() {
		t.Error("Empty HoughCircles test")
	}
	if circles.Rows() != 1 {
		t.Errorf("Invalid HoughCircles test rows: %v", circles.Rows())
	}
	if circles.Cols() < 317 || circles.Cols() > 334 {
		t.Errorf("Invalid HoughCircles test cols: %v", circles.Cols())
	}
}

func TestHoughLines(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in HoughLines test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	HoughLines(img, dest, math.Pi/180, 1, 1)
	if dest.Empty() {
		t.Error("Empty HoughLines test")
	}
	if dest.Rows() != 10411 {
		t.Errorf("Invalid HoughLines test rows: %v", dest.Rows())
	}
	if dest.Cols() != 1 {
		t.Errorf("Invalid HoughLines test cols: %v", dest.Cols())
	}

	if dest.GetFloatAt(0, 0) != 0 && dest.GetFloatAt(0, 1) != 0 {
		t.Errorf("Invalid HoughLines first test element: %v, %v", dest.GetFloatAt(0, 0), dest.GetFloatAt(0, 1))
	}

	if dest.GetFloatAt(1, 0) != 0.99483764 && dest.GetFloatAt(1, 1) != 0 {
		t.Errorf("Invalid HoughLines second test element: %v, %v", dest.GetFloatAt(1, 0), dest.GetFloatAt(1, 1))
	}

	if dest.GetFloatAt(10409, 0) != -118.246056 && dest.GetFloatAt(10409, 1) != 2 {
		t.Errorf("Invalid HoughLines penultimate test element: %v, %v", dest.GetFloatAt(10409, 0), dest.GetFloatAt(10409, 1))
	}

	if dest.GetFloatAt(10410, 0) != -118.246056 && dest.GetFloatAt(10410, 1) != 2 {
		t.Errorf("Invalid HoughLines last test element: %v, %v", dest.GetFloatAt(10410, 0), dest.GetFloatAt(10410, 1))
	}
}

func TestHoughLinesP(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in HoughLines test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	HoughLinesP(img, dest, math.Pi/180, 1, 1)
	if dest.Empty() {
		t.Error("Empty HoughLinesP test")
	}
	if dest.Rows() != 435 {
		t.Errorf("Invalid HoughLinesP test rows: %v", dest.Rows())
	}
	if dest.Cols() != 1 {
		t.Errorf("Invalid HoughLinesP test cols: %v", dest.Cols())
	}

	if dest.GetIntAt(0, 0) != 197 && dest.GetIntAt(0, 1) != 319 && dest.GetIntAt(0, 2) != 197 && dest.GetIntAt(0, 3) != 197 {
		t.Errorf("Invalid HoughLinesP first test element: %v, %v, %v, %v", dest.GetIntAt(0, 0), dest.GetIntAt(0, 1), dest.GetIntAt(0, 2), dest.GetIntAt(0, 3))
	}

	if dest.GetIntAt(1, 0) != 62 && dest.GetIntAt(1, 1) != 319 && dest.GetIntAt(1, 2) != 197 && dest.GetIntAt(1, 3) != 197 {
		t.Errorf("Invalid HoughLinesP second test element: %v, %v, %v, %v", dest.GetIntAt(1, 0), dest.GetIntAt(1, 1), dest.GetIntAt(1, 2), dest.GetIntAt(1, 3))
	}

	if dest.GetIntAt(433, 0) != 357 && dest.GetIntAt(433, 1) != 316 && dest.GetIntAt(433, 2) != 357 && dest.GetIntAt(433, 3) != 316 {
		t.Errorf("Invalid HoughLinesP penultimate test element: %v, %v, %v, %v", dest.GetIntAt(433, 0), dest.GetIntAt(433, 1), dest.GetIntAt(433, 2), dest.GetIntAt(433, 3))
	}

	if dest.GetIntAt(434, 0) != 43 && dest.GetIntAt(434, 1) != 316 && dest.GetIntAt(434, 2) != 43 && dest.GetIntAt(434, 3) != 316 {
		t.Errorf("Invalid HoughLinesP last test element: %v, %v, %v, %v", dest.GetIntAt(434, 0), dest.GetIntAt(434, 1), dest.GetIntAt(434, 2), dest.GetIntAt(434, 3))
	}
}

func TestThreshold(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in Erode test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	Threshold(img, dest, 25, 255, ThresholdBinary)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Threshold test")
	}
}
func TestAdaptiveThreshold(t *testing.T) {
	img := IMRead("images/face-detect.jpg", IMReadGrayScale)
	if img.Empty() {
		t.Error("Invalid read of Mat in AdaptiveThreshold test")
	}
	defer img.Close()

	dest := NewMat()
	defer dest.Close()

	AdaptiveThreshold(img, dest, 255, AdaptiveThresholdMean, ThresholdBinary, 11, 2)
	if dest.Empty() || img.Rows() != dest.Rows() || img.Cols() != dest.Cols() {
		t.Error("Invalid Threshold test")
	}
}

func TestDrawing(t *testing.T) {
	img := NewMatWithSize(150, 150, MatTypeCV8U)
	if img.Empty() {
		t.Error("Invalid Mat in Rectangle")
	}
	defer img.Close()

	ArrowedLine(img, image.Pt(50, 50), image.Pt(75, 75), color.RGBA{0, 0, 255, 0}, 3)
	Circle(img, image.Pt(60, 60), 20, color.RGBA{0, 0, 255, 0}, 3)
	Rectangle(img, image.Rect(50, 50, 75, 75), color.RGBA{0, 0, 255, 0}, 3)
	Line(img, image.Pt(50, 50), image.Pt(75, 75), color.RGBA{0, 0, 255, 0}, 3)

	if img.Empty() {
		t.Error("Error in Rectangle test")
	}
}

func TestGetTextSize(t *testing.T) {
	size := GetTextSize("test", FontHersheySimplex, 1.2, 1)
	if size.X != 72 {
		t.Error("Invalid text size width")
	}

	if size.Y != 26 {
		t.Error("Invalid text size height")
	}
}
func TestPutText(t *testing.T) {
	img := NewMatWithSize(150, 150, MatTypeCV8U)
	if img.Empty() {
		t.Error("Invalid Mat in IMRead")
	}
	defer img.Close()

	pt := image.Pt(10, 10)
	PutText(img, "Testing", pt, FontHersheyPlain, 1.2, color.RGBA{255, 255, 255, 0}, 2)

	if img.Empty() {
		t.Error("Error in PutText test")
	}
}

func TestResize(t *testing.T) {
	src := IMRead("images/gocvlogo.jpg", IMReadColor)
	if src.Empty() {
		t.Error("Invalid read of Mat in Resize test")
	}
	defer src.Close()

	dst := NewMat()
	defer dst.Close()

	Resize(src, dst, image.Point{}, 0.5, 0.5, InterpolationDefault)
	if dst.Cols() != 200 || dst.Rows() != 172 {
		t.Errorf("Expected dst size of 200x172 got %dx%d", dst.Cols(), dst.Rows())
	}

	Resize(src, dst, image.Pt(440, 377), 0, 0, InterpolationCubic)
	if dst.Cols() != 440 || dst.Rows() != 377 {
		t.Errorf("Expected dst size of 440x377 got %dx%d", dst.Cols(), dst.Rows())
	}
}

func TestGetRotationMatrix2D(t *testing.T) {
	type args struct {
		center image.Point
		angle  float64
		scale  float64
	}
	tests := []struct {
		name string
		args args
		want [][]float64
	}{
		{
			name: "90",
			args: args{image.Point{0, 0}, 90.0, 1.0},
			want: [][]float64{
				{6.123233995736766e-17, 1, 0},
				{-1, 6.123233995736766e-17, 0},
			},
		},
		{
			name: "45",
			args: args{image.Point{0, 0}, 45.0, 1.0},
			want: [][]float64{
				{0.7071067811865476, 0.7071067811865475, 0},
				{-0.7071067811865475, 0.7071067811865476, 0},
			},
		},
		{
			name: "0",
			args: args{image.Point{0, 0}, 0.0, 1.0},
			want: [][]float64{
				{1, 0, 0},
				{-0, 1, 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRotationMatrix2D(tt.args.center, tt.args.angle, tt.args.scale)
			for row := 0; row < got.Rows(); row++ {
				for col := 0; col < got.Cols(); col++ {
					if !floatEquals(got.GetDoubleAt(row, col), tt.want[row][col]) {
						t.Errorf("GetRotationMatrix2D() = %v, want %v at row:%v col:%v", got.GetDoubleAt(row, col), tt.want[row][col], row, col)
					}
				}
			}
		})
	}
}

func TestWarpAffine(t *testing.T) {
	src := NewMatWithSize(256, 256, MatTypeCV8UC1)
	rot := GetRotationMatrix2D(image.Point{0, 0}, 1.0, 1.0)
	dst := src.Clone()

	WarpAffine(src, dst, rot, image.Point{256, 256})
	result := Norm(dst, NormL2)
	if result != 0.0 {
		t.Errorf("WarpAffine() = %v, want %v", result, 0.0)
	}
	src = IMRead("images/gocvlogo.jpg", IMReadUnchanged)
	dst = src.Clone()
	WarpAffine(src, dst, rot, image.Point{343, 400})
	result = Norm(dst, NormL2)

	if !floatEquals(round(result, 0.05), round(111111.05, 0.05)) {
		t.Errorf("WarpAffine() = %v, want %v", round(result, 0.05), round(111111.05, 0.05))
	}
}

func TestWarpAffineWithParams(t *testing.T) {
	src := NewMatWithSize(256, 256, MatTypeCV8UC1)
	rot := GetRotationMatrix2D(image.Point{0, 0}, 1.0, 1.0)
	dst := src.Clone()

	WarpAffineWithParams(src, dst, rot, image.Point{256, 256}, InterpolationLinear, BorderConstant, color.RGBA{0, 0, 0, 0})
	result := Norm(dst, NormL2)
	if !floatEquals(result, 0.0) {
		t.Errorf("WarpAffineWithParams() = %v, want %v", result, 0.0)
	}

	src = IMRead("images/gocvlogo.jpg", IMReadUnchanged)
	dst = src.Clone()
	WarpAffineWithParams(src, dst, rot, image.Point{343, 400}, InterpolationLinear, BorderConstant, color.RGBA{0, 0, 0, 0})
	result = Norm(dst, NormL2)
	if !floatEquals(round(result, 0.05), round(111111.05, 0.05)) {
		t.Errorf("WarpAffine() = %v, want %v", round(result, 0.05), round(111111.05, 0.05))
	}
}

func TestApplyColorMap(t *testing.T) {
	type args struct {
		colormapType ColormapTypes
		want         float64
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "COLORMAP_AUTUMN", args: args{colormapType: ColormapAutumn, want: 118090.29593069873}},
		{name: "COLORMAP_BONE", args: args{colormapType: ColormapBone, want: 122067.44213343704}},
		{name: "COLORMAP_JET", args: args{colormapType: ColormapJet, want: 98220.64722857409}},
		{name: "COLORMAP_WINTER", args: args{colormapType: ColormapWinter, want: 94279.52859449394}},
		{name: "COLORMAP_RAINBOW", args: args{colormapType: ColormapRainbow, want: 92591.40608069411}},
		{name: "COLORMAP_OCEAN", args: args{colormapType: ColormapOcean, want: 106444.16919681415}},
		{name: "COLORMAP_SUMMER", args: args{colormapType: ColormapSummer, want: 114434.44957703952}},
		{name: "COLORMAP_SPRING", args: args{colormapType: ColormapSpring, want: 123557.60209715953}},
		{name: "COLORMAP_COOL", args: args{colormapType: ColormapCool, want: 123557.60209715953}},
		{name: "COLORMAP_HSV", args: args{colormapType: ColormapHsv, want: 107679.25179903508}},
		{name: "COLORMAP_PINK", args: args{colormapType: ColormapPink, want: 136043.97287274434}},
		{name: "COLORMAP_HOT", args: args{colormapType: ColormapHot, want: 124941.02475968412}},
		{name: "COLORMAP_PARULA", args: args{colormapType: ColormapParula, want: 111483.33555738274}},
	}
	src := IMRead("images/gocvlogo.jpg", IMReadGrayScale)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := src.Clone()
			ApplyColorMap(src, dst, tt.args.colormapType)
			result := Norm(dst, NormL2)
			if !floatEquals(result, tt.args.want) {
				t.Errorf("TestApplyColorMap() = %v, want %v", result, tt.args.want)
			}
		})
	}
}

func TestApplyCustomColorMap(t *testing.T) {
	src := IMRead("images/gocvlogo.jpg", IMReadGrayScale)
	customColorMap := NewMatWithSize(256, 1, MatTypeCV8UC1)

	dst := src.Clone()
	ApplyCustomColorMap(src, dst, customColorMap)
	result := Norm(dst, NormL2)
	if !floatEquals(result, 0.0) {
		t.Errorf("TestApplyCustomColorMap() = %v, want %v", result, 0.0)
	}
}

func TestGetPerspectiveTransform(t *testing.T) {
	src := []image.Point{
		image.Pt(0, 0),
		image.Pt(10, 5),
		image.Pt(10, 10),
		image.Pt(5, 10),
	}
	dst := []image.Point{
		image.Pt(0, 0),
		image.Pt(10, 0),
		image.Pt(10, 10),
		image.Pt(0, 10),
	}

	m := GetPerspectiveTransform(src, dst)

	if m.Cols() != 3 {
		t.Errorf("TestWarpPerspective(): unexpected cols = %v, want = %v", m.Cols(), 3)
	}
	if m.Rows() != 3 {
		t.Errorf("TestWarpPerspective(): unexpected rows = %v, want = %v", m.Rows(), 3)
	}
}

func TestWarpPerspective(t *testing.T) {
	img := IMRead("images/gocvlogo.jpg", IMReadUnchanged)
	defer img.Close()

	w := img.Cols()
	h := img.Rows()

	s := []image.Point{
		image.Pt(0, 0),
		image.Pt(10, 5),
		image.Pt(10, 10),
		image.Pt(5, 10),
	}
	d := []image.Point{
		image.Pt(0, 0),
		image.Pt(10, 0),
		image.Pt(10, 10),
		image.Pt(0, 10),
	}
	m := GetPerspectiveTransform(s, d)

	dst := NewMat()
	defer dst.Close()

	WarpPerspective(img, dst, m, image.Pt(w, h))

	if dst.Cols() != w {
		t.Errorf("TestWarpPerspective(): unexpected cols = %v, want = %v", dst.Cols(), w)
	}

	if dst.Rows() != h {
		t.Errorf("TestWarpPerspective(): unexpected rows = %v, want = %v", dst.Rows(), h)
	}
}

func TestDrawContours(t *testing.T) {
	img := NewMatWithSize(100, 200, MatTypeCV8UC1)
	defer img.Close()

	// Draw rectangle
	white := color.RGBA{255, 255, 255, 255}
	Rectangle(img, image.Rect(125, 25, 175, 75), white, 1)

	contours := FindContours(img, RetrievalExternal, ChainApproxSimple)

	if v := img.GetUCharAt(23, 123); v != 0 {
		t.Errorf("TestDrawContours(): wrong pixel value = %v, want = %v", v, 0)
	}
	if v := img.GetUCharAt(25, 125); v != 206 {
		t.Errorf("TestDrawContours(): wrong pixel value = %v, want = %v", v, 206)
	}

	DrawContours(img, contours, -1, white, 2)

	// contour should be drawn with thickness = 2
	if v := img.GetUCharAt(24, 124); v != 255 {
		t.Errorf("TestDrawContours(): contour has not been drawn (value = %v, want = %v)", v, 255)
	}
	if v := img.GetUCharAt(25, 125); v != 255 {
		t.Errorf("TestDrawContours(): contour has not been drawn (value = %v, want = %v)", v, 255)
	}
}
