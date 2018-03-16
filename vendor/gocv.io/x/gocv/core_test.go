package gocv

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestMat(t *testing.T) {
	mat := NewMat()
	if !mat.Empty() {
		t.Error("New Mat should be empty")
	}
}

func TestMatWithSize(t *testing.T) {
	mat := NewMatWithSize(101, 102, MatTypeCV8U)
	if mat.Empty() {
		t.Error("NewMatWithSize should not be empty")
	}

	if mat.Rows() != 101 {
		t.Errorf("NewMatWithSize incorrect row count: %v\n", mat.Rows())
	}

	if mat.Cols() != 102 {
		t.Errorf("NewMatWithSize incorrect col count: %v\n", mat.Cols())
	}

	if mat.Channels() != 1 {
		t.Errorf("NewMatWithSize incorrect channels count: %v\n", mat.Channels())
	}

	if mat.Type() != 0 {
		t.Errorf("NewMatWithSize incorrect type: %v\n", mat.Type())
	}
}

func TestMatClone(t *testing.T) {
	mat := NewMatWithSize(101, 102, MatTypeCV8U)
	clone := mat.Clone()
	if clone.Rows() != 101 {
		t.Errorf("Mat clone incorrect row count: %v\n", clone.Rows())
	}

	if clone.Cols() != 102 {
		t.Errorf("Mat clone incorrect col count: %v\n", clone.Cols())
	}
}

func TestMatCopyTo(t *testing.T) {
	mat := NewMatWithSize(101, 102, MatTypeCV8U)
	defer mat.Close()
	copy := NewMat()
	defer copy.Close()

	mat.CopyTo(copy)
	if copy.Rows() != 101 {
		t.Errorf("Mat copy incorrect row count: %v\n", copy.Rows())
	}

	if copy.Cols() != 102 {
		t.Errorf("Mat copy incorrect col count: %v\n", copy.Cols())
	}
}

func TestMatCopyToWithMask(t *testing.T) {
	mat := NewMatWithSize(101, 102, MatTypeCV8U)
	defer mat.Close()
	mask := NewMatWithSize(101, 102, MatTypeCV8U)
	defer mask.Close()
	diff := NewMat()
	defer diff.Close()

	mat.SetUCharAt(0, 0, 255)
	mat.SetUCharAt(0, 1, 255)

	mask.SetUCharAt(0, 0, 255)

	copy := NewMat()
	defer copy.Close()

	mat.CopyToWithMask(copy, mask)
	if copy.Rows() != 101 {
		t.Errorf("Mat copy incorrect row count: %v\n", copy.Rows())
	}

	if copy.Cols() != 102 {
		t.Errorf("Mat copy incorrect col count: %v\n", copy.Cols())
	}

	if copy.GetUCharAt(0, 0) != 255 || copy.GetUCharAt(0, 1) != 0 {
		t.Errorf("Mask failed to apply to source image")
	}

	Compare(mat, copy, diff, CompareEQ)
	if CountNonZero(diff) == 0 {
		t.Errorf("Mat CopyToWithMask incorrect diff: %v\n", CountNonZero(diff))
	}
}

func TestMatToBytes(t *testing.T) {
	mat := NewMatWithSize(101, 102, MatTypeCV8U)
	b := mat.ToBytes()
	if len(b) != 101*102 {
		t.Errorf("Mat bytes incorrect length: %v\n", len(b))
	}

	copy := NewMatFromBytes(101, 102, MatTypeCV8U, b)
	if copy.Rows() != 101 {
		t.Errorf("Mat from bytes incorrect row count: %v\n", copy.Rows())
	}
	if copy.Cols() != 102 {
		t.Errorf("Mat region incorrect col count: %v\n", copy.Cols())
	}

	mat = NewMatWithSize(101, 102, MatTypeCV16S)
	b = mat.ToBytes()
	if len(b) != 101*102*2 {
		t.Errorf("Mat bytes incorrect length: %v\n", len(b))
	}

	mat = NewMatFromScalar(NewScalar(255.0, 105.0, 180.0, 0.0), MatTypeCV8UC3)
	b = mat.ToBytes()
	if len(b) != 3 {
		t.Errorf("Mat bytes incorrect length: %v\n", len(b))
	}
	if bytes.Compare(b, []byte{255, 105, 180}) != 0 {
		t.Errorf("Mat bytes unexpected values: %v\n", b)
	}
}

func TestMatRegion(t *testing.T) {
	mat := NewMatWithSize(100, 100, MatTypeCV8U)
	region := mat.Region(image.Rect(20, 25, 80, 75))
	if region.Rows() != 50 {
		t.Errorf("Mat region incorrect row count: %v\n", region.Rows())
	}

	if region.Cols() != 60 {
		t.Errorf("Mat region incorrect col count: %v\n", region.Cols())
	}
}

func TestMatReshape(t *testing.T) {
	mat := NewMatWithSize(100, 100, MatTypeCV8UC4)
	defer mat.Close()

	r := mat.Reshape(1, 1)
	if r.Rows() != 1 {
		t.Errorf("Mat reshape incorrect row count: %v\n", r.Rows())
	}

	if r.Cols() != 40000 {
		t.Errorf("Mat reshape incorrect col count: %v\n", r.Cols())
	}
}

func TestMatConvert(t *testing.T) {
	src := NewMatWithSize(100, 100, MatTypeCV32F)
	dst := NewMat()
	src.ConvertTo(dst, MatTypeCV16S)
	if dst.Empty() {
		t.Error("TestConvert dst should not be empty.")
	}
}

func TestMatConvertFp16(t *testing.T) {
	src := NewMatWithSize(100, 100, MatTypeCV32F)
	dst := src.ConvertFp16()
	if dst.Empty() {
		t.Error("TestConvertFp16 dst should not be empty.")
	}
}

func TestMatMean(t *testing.T) {
	mat := NewMatWithSize(100, 100, MatTypeCV8U)
	mean := mat.Mean()
	if mean.Val1 != 0 {
		t.Errorf("Mat Mean incorrect Val1")
	}
}

func TestLUT(t *testing.T) {
	src := IMRead("images/gocvlogo.jpg", IMReadColor)
	if src.Empty() {
		t.Error("Invalid read of Source Mat in LUT test")
	}
	defer src.Close()

	lut := IMRead("images/lut.png", IMReadColor)
	if lut.Empty() {
		t.Error("Invalid read of LUT Mat in LUT test")
	}
	defer lut.Close()

	dst := NewMat()
	defer dst.Close()

	LUT(src, lut, dst)
	if dst.Cols() != 400 || dst.Rows() != 343 {
		t.Errorf("Expected dst size of 200x172 got %dx%d", dst.Cols(), dst.Rows())
	}
}

func TestMatAccessors(t *testing.T) {
	mat := NewMatWithSize(101, 102, MatTypeCV8U)
	if mat.GetUCharAt(50, 50) != 0 {
		t.Errorf("GetUCharAt incorrect value: %v\n", mat.GetUCharAt(50, 50))
	}
	if mat.GetUCharAt3(50, 50, 0) != 0 {
		t.Errorf("GetUCharAt3 incorrect value: %v\n", mat.GetUCharAt3(50, 50, 0))
	}
	mat.Close()

	mat = NewMatWithSize(101, 102, MatTypeCV8S)
	if mat.GetSCharAt(50, 50) != 0 {
		t.Errorf("GetSCharAt incorrect value: %v\n", mat.GetSCharAt(50, 50))
	}
	if mat.GetSCharAt3(50, 50, 0) != 0 {
		t.Errorf("GetSCharAt3 incorrect value: %v\n", mat.GetSCharAt3(50, 50, 0))
	}
	mat.Close()

	mat = NewMatWithSize(101, 102, MatTypeCV16S)
	if mat.GetShortAt(50, 50) != 0 {
		t.Errorf("GetShortAt incorrect value: %v\n", mat.GetShortAt(50, 50))
	}
	if mat.GetShortAt3(50, 50, 0) != 0 {
		t.Errorf("GetShortAt3 incorrect value: %v\n", mat.GetShortAt3(50, 50, 0))
	}
	mat.Close()

	mat = NewMatWithSize(101, 102, MatTypeCV32S)
	if mat.GetIntAt(50, 50) != 0 {
		t.Errorf("GetIntAt incorrect value: %v\n", mat.GetIntAt(50, 50))
	}
	if mat.GetIntAt3(50, 50, 0) != 0 {
		t.Errorf("GetIntAt3 incorrect value: %v\n", mat.GetIntAt3(50, 50, 0))
	}
	mat.Close()

	mat = NewMatWithSize(101, 102, MatTypeCV32F)
	if mat.GetFloatAt(50, 50) != 0.0 {
		t.Errorf("GetFloatAt incorrect value: %v\n", mat.GetFloatAt(50, 50))
	}
	if mat.GetFloatAt3(50, 50, 0) != 0.0 {
		t.Errorf("GetFloatAt3 incorrect value: %v\n", mat.GetFloatAt3(50, 50, 0))
	}
	mat.Close()

	mat = NewMatWithSize(101, 102, MatTypeCV64F)
	if mat.GetDoubleAt(50, 50) != 0.0 {
		t.Errorf("GetDoubleAt incorrect value: %v\n", mat.GetDoubleAt(50, 50))
	}
	if mat.GetDoubleAt3(50, 50, 0) != 0.0 {
		t.Errorf("GetDoubleAt3 incorrect value: %v\n", mat.GetDoubleAt3(50, 50, 0))
	}
	mat.Close()
}

func TestMatMutators(t *testing.T) {
	t.Run("SetUCharAt", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV8U)
		mat.SetUCharAt(50, 50, 25)
		if mat.GetUCharAt(50, 50) != 25 {
			t.Errorf("SetUCharAt incorrect value: %v\n", mat.GetUCharAt(50, 50))
		}
		mat.Close()
	})
	t.Run("SetUCharAt3", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV8U)
		mat.SetUCharAt3(50, 50, 0, 25)
		if mat.GetUCharAt3(50, 50, 0) != 25 {
			t.Errorf("SetUCharAt3 incorrect value: %v\n", mat.GetUCharAt3(50, 50, 0))
		}
		mat.Close()
	})
	t.Run("SetSCharAt", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV8S)
		mat.SetSCharAt(50, 50, 25)
		if mat.GetSCharAt(50, 50) != 25 {
			t.Errorf("SetSCharAt incorrect value: %v\n", mat.GetSCharAt(50, 50))
		}
		mat.Close()
	})
	t.Run("SetSCharAt3", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV8S)
		mat.SetSCharAt3(50, 50, 0, 25)
		if mat.GetSCharAt3(50, 50, 0) != 25 {
			t.Errorf("SetSCharAt3 incorrect value: %v\n", mat.GetSCharAt3(50, 50, 0))
		}
		mat.Close()
	})
	t.Run("SetShortAt", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV16S)
		mat.SetShortAt(50, 50, 25)
		if mat.GetShortAt(50, 50) != 25 {
			t.Errorf("SetShortAt incorrect value: %v\n", mat.GetShortAt(50, 50))
		}
		mat.Close()
	})
	t.Run("SetShortAt3", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV16S)
		mat.SetShortAt3(50, 50, 0, 25)
		if mat.GetShortAt3(50, 50, 0) != 25 {
			t.Errorf("SetShortAt3 incorrect value: %v\n", mat.GetShortAt3(50, 50, 0))
		}
		mat.Close()
	})
	t.Run("SetIntAt", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV32S)
		mat.SetIntAt(50, 50, 25)
		if mat.GetIntAt(50, 50) != 25 {
			t.Errorf("SetIntAt incorrect value: %v\n", mat.GetIntAt(50, 50))
		}
		mat.Close()
	})
	t.Run("SetIntAt3", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV32S)
		mat.SetIntAt3(50, 50, 0, 25)
		if mat.GetIntAt3(50, 50, 0) != 25 {
			t.Errorf("SetIntAt3 incorrect value: %v\n", mat.GetIntAt3(50, 50, 0))
		}
		mat.Close()
	})
	t.Run("SetFloatAt", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV32F)
		mat.SetFloatAt(50, 50, 25.0)
		if mat.GetFloatAt(50, 50) != 25 {
			t.Errorf("SetFloatAt incorrect value: %v\n", mat.GetFloatAt(50, 50))
		}
		mat.Close()
	})
	t.Run("SetFloatAt3", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV32F)
		mat.SetFloatAt3(50, 50, 0, 25.0)
		if mat.GetFloatAt3(50, 50, 0) != 25 {
			t.Errorf("SetFloatAt incorrect value: %v\n", mat.GetFloatAt3(50, 50, 0))
		}
		mat.Close()
	})
	t.Run("SetDoubleAt", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV64F)
		mat.SetDoubleAt(50, 50, 25.0)
		if mat.GetDoubleAt(50, 50) != 25.0 {
			t.Errorf("SetDoubleAt incorrect value: %v\n", mat.GetDoubleAt(50, 50))
		}
		mat.Close()
	})
	t.Run("SetDoubleAt3", func(t *testing.T) {
		mat := NewMatWithSize(101, 102, MatTypeCV64F)
		mat.SetDoubleAt3(50, 50, 0, 25.0)
		if mat.GetDoubleAt3(50, 50, 0) != 25.0 {
			t.Errorf("SetDoubleAt3 incorrect value: %v\n", mat.GetDoubleAt3(50, 50, 0))
		}
		mat.Close()
	})
}

func TestMatAbsDiff(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat2 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat3 := NewMat()
	AbsDiff(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatAbsDiff dest mat3 should not be empty.")
	}
}

func TestMatAdd(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat2 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat3 := NewMat()
	Add(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatAdd dest mat3 should not be empty.")
	}
}

func TestMatAddWeighted(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat2 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat3 := NewMat()
	AddWeighted(mat1, 2.0, mat2, 3.0, 4.0, mat3)
	if mat3.Empty() {
		t.Error("TestMatAddWeighted dest mat3 should not be empty.")
	}
}

func TestMatBitwiseOperations(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat2 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat3 := NewMat()
	BitwiseAnd(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatBitwiseAnd dest mat3 should not be empty.")
	}

	BitwiseOr(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatBitwiseOr dest mat3 should not be empty.")
	}

	BitwiseXor(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatBitwiseXor dest mat3 should not be empty.")
	}

	BitwiseNot(mat1, mat3)
	if mat3.Empty() {
		t.Error("TestMatBitwiseNot dest mat3 should not be empty.")
	}

}

func TestMatInRange(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	lb := NewMatFromScalar(NewScalar(20.0, 100.0, 100.0, 0.0), MatTypeCV8U)
	ub := NewMatFromScalar(NewScalar(20.0, 100.0, 100.0, 0.0), MatTypeCV8U)
	dst := NewMat()
	InRange(mat1, lb, ub, dst)
	if dst.Empty() {
		t.Error("TestMatAddWeighted dest mat3 should not be empty.")
	}
}

func TestMatDCT(t *testing.T) {
	src := NewMatWithSize(64, 64, MatTypeCV32F)
	dst := NewMat()

	DCT(src, dst, DftForward)
	if dst.Empty() {
		t.Error("TestMatDCT dst should not be empty.")
	}
}

func TestMatDFT(t *testing.T) {
	src := NewMatWithSize(101, 102, MatTypeCV32F)
	dst := NewMat()

	m := GetOptimalDFTSize(101)
	n := GetOptimalDFTSize(102)
	if m != 108 {
		t.Errorf("TestMatOptimalDFT dst error: %d", m)
	}

	if n != 108 {
		t.Errorf("TestMatOptimalDFT dst error: %d", n)
	}

	DFT(src, dst, DftForward)
	if dst.Empty() {
		t.Error("TestMatDFT dst should not be empty.")
	}
}

func TestMatDivide(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat2 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat3 := NewMat()
	Divide(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatDivide dest mat3 should not be empty.")
	}
}

func TestMatMerge(t *testing.T) {
	src := NewMatWithSize(101, 102, MatTypeCV8U)
	src2 := NewMatWithSize(101, 102, MatTypeCV8U)
	src3 := NewMatWithSize(101, 102, MatTypeCV8U)
	dst := NewMat()
	Merge([]Mat{src, src2, src3}, dst)
	if dst.Empty() {
		t.Error("TestMatMerge dst should not be empty.")
	}
}

func TestMatMultiply(t *testing.T) {
	mat1 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat2 := NewMatWithSize(101, 102, MatTypeCV8U)
	mat3 := NewMat()
	Multiply(mat1, mat2, mat3)
	if mat3.Empty() {
		t.Error("TestMatMultiply dest mat3 should not be empty.")
	}
}

func TestMatNormalize(t *testing.T) {
	src := NewMatWithSize(101, 102, MatTypeCV8U)
	dst := NewMat()
	Normalize(src, dst, 0.0, 255.0, NormMixMax)
	if dst.Empty() {
		t.Error("TestMatNormalize dst should not be empty.")
	}
}

func TestMatSplit(t *testing.T) {
	src := IMRead("images/face.jpg", 1)
	chans := Split(src)
	if len(chans) != src.Channels() {
		t.Error("Split Channel count differs")
	}
	dst := NewMat()
	Merge(chans, dst)
	diff := NewMat()
	AbsDiff(src, dst, diff)
	sum := diff.Sum()
	if sum.Val1 != 0 || sum.Val2 != 0 || sum.Val3 != 0 {
		t.Error("Split/Merged images differ")
	}
}

func TestMatSubtract(t *testing.T) {
	src1 := IMRead("images/lut.png", 1)
	src2 := IMRead("images/lut.png", 1)
	dst := NewMat()
	Subtract(src1, src2, dst)
	sum := dst.Sum()
	if sum.Val1 != 0 || sum.Val2 != 0 || sum.Val3 != 0 {
		t.Error("Sum of Subtracting equal images is not 0")
	}
}
func TestMatPow(t *testing.T) {
	src := NewMatWithSize(101, 102, MatTypeCV8U)
	dst := NewMat()
	power := 2.0
	Pow(src, power, dst)

	if dst.Empty() {
		t.Error("TestMatPow dest should not be empty.")
	}
}

func TestMatSum(t *testing.T) {
	src := NewMatFromScalar(NewScalar(1, 2, 3, 4), MatTypeCV8UC4)
	sum := src.Sum()
	if sum.Val1 != 1 || sum.Val2 != 2 || sum.Val3 != 3 || sum.Val4 != 4 {
		t.Error("Sum values do not match constructor")
	}
}

func TestTermCriteria(t *testing.T) {
	tc := NewTermCriteria(Count, 50, 2.0)
	if tc.p == nil {
		t.Error("TermCriteria has invalid value")
	}
}

func TestScalar(t *testing.T) {
	s := NewScalar(127.0, 255.0, 64.0, 0.0)
	if s.Val1 != 127.0 || s.Val2 != 255.0 || s.Val3 != 64.0 || s.Val4 != 0.0 {
		t.Error("Scalar has invalid value")
	}
}

func TestToCPoints(t *testing.T) {
	points := []image.Point{
		image.Pt(0, 0),
		image.Pt(1, 1),
	}

	cPoints := toCPoints(points)

	if int(cPoints.length) != len(points) {
		t.Error("Invalid C Points length")
	}
}

func TestMatBatchDistance(t *testing.T) {
	src1 := NewMatWithSize(100, 100, MatTypeCV8U)
	src2 := NewMatWithSize(100, 100, MatTypeCV8U)
	mask := NewMatWithSize(100, 100, MatTypeCV8U)
	dist := NewMat()
	nidx := NewMat()
	BatchDistance(src1, src2, dist, -1, nidx, NormL2, 15, mask, 0, false)
	if dist.Empty() {
		t.Error("TestBatchDistance dst should not be empty.")
	}
	src1.Close()
	src2.Close()
	mask.Close()
	dist.Close()
	nidx.Close()
}

func TestMatBorderInterpolate(t *testing.T) {
	n := BorderInterpolate(1, 5, 1)
	if n == 0 {
		t.Error("TestBorderInterpolate dst should not be 0.")
	}
}

func TestMatCalcCovarMatrix(t *testing.T) {
	samples := NewMatWithSize(10, 10, MatTypeCV32F)
	covar := NewMat()
	mean := NewMat()
	CalcCovarMatrix(samples, covar, mean, CovarRows, MatTypeCV64F)
	if covar.Empty() {
		t.Error("TestCalcCovarMatrix dst should not be empty.")
	}
	samples.Close()
	covar.Close()
	mean.Close()
}

func TestMatCartToPolar(t *testing.T) {
	x := NewMatWithSize(100, 100, MatTypeCV32F)
	y := NewMatWithSize(100, 100, MatTypeCV32F)
	magnitude := NewMat()
	angle := NewMat()
	CartToPolar(x, y, magnitude, angle, false)
	if magnitude.Empty() || angle.Empty() {
		t.Error("TestCartToPolar neither magnitude nor angle should be empty.")
	}
	x.Close()
	y.Close()
	magnitude.Close()
	angle.Close()
}

func TestMatCompleteSymm(t *testing.T) {
	src := NewMatWithSize(100, 100, MatTypeCV32F)
	CompleteSymm(src, false)
	if src.Empty() {
		t.Error("TestCompleteSymm src should not be empty.")
	}
	src.Close()
}

func TestMatConvertScaleAbs(t *testing.T) {
	src := NewMatWithSize(100, 100, MatTypeCV32F)
	dst := NewMat()
	ConvertScaleAbs(src, dst, 1, 0)
	if dst.Empty() {
		t.Error("TestConvertScaleAbs dst should not be empty.")
	}
	src.Close()
	dst.Close()
}

func TestMatCopyMakeBorder(t *testing.T) {
	src := NewMatWithSize(100, 100, MatTypeCV32F)
	dst := NewMat()
	CopyMakeBorder(src, dst, 10, 10, 10, 10, BorderReflect, color.RGBA{0, 0, 0, 0})
	if dst.Empty() {
		t.Error("TestCopyMakeBorder dst should not be empty.")
	}
	src.Close()
	dst.Close()
}

func TestMatEigen(t *testing.T) {
	src := NewMatWithSize(10, 10, MatTypeCV32F)
	eigenvalues := NewMat()
	eigenvectors := NewMat()
	Eigen(src, eigenvalues, eigenvectors)
	if eigenvectors.Empty() || eigenvalues.Empty() {
		t.Error("TestEigen should not have empty eigenvectors or eigenvalues.")
	}
	src.Close()
	eigenvectors.Close()
	eigenvalues.Close()
}

func TestMatExp(t *testing.T) {
	src := NewMatWithSize(10, 10, MatTypeCV32F)
	dst := NewMat()
	Exp(src, dst)
	if dst.Empty() {
		t.Error("TestExp dst should not be empty.")
	}
	src.Close()
	dst.Close()
}

func TestMatExtractChannel(t *testing.T) {
	src := NewMatWithSize(10, 10, MatTypeCV32F+MatChannels3)
	dst := NewMat()
	ExtractChannel(src, dst, 1)
	if dst.Empty() {
		t.Error("TestExtractChannel dst should not be empty.")
	}
	src.Close()
	dst.Close()
}
