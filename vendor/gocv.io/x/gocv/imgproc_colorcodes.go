package gocv

// ColorConversionCode is a color conversion code used on Mat.
//
// For further details, please see:
// http://docs.opencv.org/master/d7/d1b/group__imgproc__misc.html#ga4e0972be5de079fed4e3a10e24ef5ef0
//
type ColorConversionCode int

const (
	// ColorBGRToBGRA adds alpha channel to BGR image.
	ColorBGRToBGRA ColorConversionCode = 0

	// ColorBGRAToBGR removes alpha channel from BGR image.
	ColorBGRAToBGR = 1

	// ColorBGRToRGBA converts from BGR to RGB with alpha channel.
	ColorBGRToRGBA = 2

	// ColorRGBAToBGR converts from RGB with alpha to BGR color space.
	ColorRGBAToBGR = 3

	// ColorBGRToRGB converts from BGR to RGB without alpha channel.
	ColorBGRToRGB = 4

	// ColorBGRAToRGBA converts from BGR with alpha channel
	// to RGB with alpha channel.
	ColorBGRAToRGBA = 5

	// ColorBGRToGray converts from BGR to grayscale.
	ColorBGRToGray = 6

	// ColorRGBToGray converts from RGB to grayscale.
	ColorRGBToGray = 7

	// ColorGrayToBGR converts from grayscale to BGR.
	ColorGrayToBGR = 8

	// ColorGrayToBGRA converts from grayscale to BGR with alpha channel.
	ColorGrayToBGRA = 9

	// ColorBGRAToGray converts from BGR with alpha channel to grayscale.
	ColorBGRAToGray = 10

	// ColorRGBAToGray converts from RGB with alpha channel to grayscale.
	ColorRGBAToGray = 11

	// ColorBGRToBGR565 converts from BGR to BGR565 (16-bit images).
	ColorBGRToBGR565 = 12

	// ColorRGBToBGR565 converts from RGB to BGR565 (16-bit images).
	ColorRGBToBGR565 = 13

	// ColorBGR565ToBGR converts from BGR565 (16-bit images) to BGR.
	ColorBGR565ToBGR = 14

	// ColorBGR565ToRGB converts from BGR565 (16-bit images) to RGB.
	ColorBGR565ToRGB = 15

	// ColorBGRAToBGR565 converts from BGRA (with alpha channel)
	// to BGR565 (16-bit images).
	ColorBGRAToBGR565 = 16

	// ColorRGBAToBGR565 converts from RGBA (with alpha channel)
	// to BGR565 (16-bit images).
	ColorRGBAToBGR565 = 17

	// ColorBGR565ToBGRA converts from BGR565 (16-bit images)
	// to BGRA (with alpha channel).
	ColorBGR565ToBGRA = 18

	// ColorBGR565ToRGBA converts from BGR565 (16-bit images)
	// to RGBA (with alpha channel).
	ColorBGR565ToRGBA = 19

	// ColorGrayToBGR565 converts from grayscale
	// to BGR565 (16-bit images).
	ColorGrayToBGR565 = 20

	// ColorBGR565ToGray converts from BGR565 (16-bit images)
	// to grayscale.
	ColorBGR565ToGray = 21

	// ColorBGRToBGR555 converts from BGR to BGR555 (16-bit images).
	ColorBGRToBGR555 = 22

	// ColorRGBToBGR555 converts from RGB to BGR555 (16-bit images).
	ColorRGBToBGR555 = 23

	// ColorBGR555ToBGR converts from BGR555 (16-bit images) to BGR.
	ColorBGR555ToBGR = 24

	// ColorBGR555ToRGB converts from BGR555 (16-bit images) to RGB.
	ColorBGR555ToRGB = 25

	// ColorBGRAToBGR555 converts from BGRA (with alpha channel)
	// to BGR555 (16-bit images).
	ColorBGRAToBGR555 = 26

	// ColorRGBAToBGR555 converts from RGBA (with alpha channel)
	// to BGR555 (16-bit images).
	ColorRGBAToBGR555 = 27

	// ColorBGR555ToBGRA converts from BGR555 (16-bit images)
	// to BGRA (with alpha channel).
	ColorBGR555ToBGRA = 28

	// ColorBGR555ToRGBA converts from BGR555 (16-bit images)
	// to RGBA (with alpha channel).
	ColorBGR555ToRGBA = 29

	// ColorGrayToBGR555 converts from grayscale to BGR555 (16-bit images).
	ColorGrayToBGR555 = 30

	// ColorBGR555ToGRAY converts from BGR555 (16-bit images) to grayscale.
	ColorBGR555ToGRAY = 31

	// ColorBGRToXYZ converts from BGR to CIE XYZ.
	ColorBGRToXYZ = 32

	// ColorRGBToXYZ converts from RGB to CIE XYZ.
	ColorRGBToXYZ = 33

	// ColorXYZToBGR converts from CIE XYZ to BGR.
	ColorXYZToBGR = 34

	// ColorXYZToRGB converts from CIE XYZ to RGB.
	ColorXYZToRGB = 35

	// ColorBGRToYCrCb converts from BGR to luma-chroma (aka YCC).
	ColorBGRToYCrCb = 36

	// ColorRGBToYCrCb converts from RGB to luma-chroma (aka YCC).
	ColorRGBToYCrCb = 37

	// ColorYCrCbToBGR converts from luma-chroma (aka YCC) to BGR.
	ColorYCrCbToBGR = 38

	// ColorYCrCbToRGB converts from luma-chroma (aka YCC) to RGB.
	ColorYCrCbToRGB = 39

	// ColorBGRToHSV converts from BGR to HSV (hue saturation value).
	ColorBGRToHSV = 40

	// ColorRGBToHSV converts from RGB to HSV (hue saturation value).
	ColorRGBToHSV = 41

	// ColorBGRToLab converts from BGR to CIE Lab.
	ColorBGRToLab = 44

	// ColorRGBToLab converts from RGB to CIE Lab.
	ColorRGBToLab = 45

	// ColorBGRToLuv converts from BGR to CIE Luv.
	ColorBGRToLuv = 50

	// ColorRGBToLuv converts from RGB to CIE Luv.
	ColorRGBToLuv = 51

	// ColorBGRToHLS converts from BGR to HLS (hue lightness saturation).
	ColorBGRToHLS = 52

	// ColorRGBToHLS converts from RGB to HLS (hue lightness saturation).
	ColorRGBToHLS = 53

	// ColorHSVToBGR converts from HSV (hue saturation value) to BGR.
	ColorHSVToBGR = 54

	// ColorHSVToRGB converts from HSV (hue saturation value) to RGB.
	ColorHSVToRGB = 55

	// ColorLabToBGR converts from CIE Lab to BGR.
	ColorLabToBGR = 56

	// ColorLabToRGB converts from CIE Lab to RGB.
	ColorLabToRGB = 57

	// ColorLuvToBGR converts from CIE Luv to BGR.
	ColorLuvToBGR = 58

	// ColorLuvToRGB converts from CIE Luv to RGB.
	ColorLuvToRGB = 59

	// ColorHLSToBGR converts from HLS (hue lightness saturation) to BGR.
	ColorHLSToBGR = 60

	// ColorHLSToRGB converts from HLS (hue lightness saturation) to RGB.
	ColorHLSToRGB = 61

	// ColorBGRToHSVFull converts from BGR to HSV (hue saturation value) full.
	ColorBGRToHSVFull = 66

	// ColorRGBToHSVFull converts from RGB to HSV (hue saturation value) full.
	ColorRGBToHSVFull = 67

	// ColorBGRToHLSFull converts from BGR to HLS (hue lightness saturation) full.
	ColorBGRToHLSFull = 68

	// ColorRGBToHLSFull converts from RGB to HLS (hue lightness saturation) full.
	ColorRGBToHLSFull = 69

	// ColorHSVToBGRFull converts from HSV (hue saturation value) to BGR full.
	ColorHSVToBGRFull = 70

	// ColorHSVToRGBFull converts from HSV (hue saturation value) to RGB full.
	ColorHSVToRGBFull = 71

	// ColorHLSToBGRFull converts from HLS (hue lightness saturation) to BGR full.
	ColorHLSToBGRFull = 72

	// ColorHLSToRGBFull converts from HLS (hue lightness saturation) to RGB full.
	ColorHLSToRGBFull = 73

	// ColorLBGRToLab converts from LBGR to CIE Lab.
	ColorLBGRToLab = 74

	// ColorLRGBToLab converts from LRGB to CIE Lab.
	ColorLRGBToLab = 75

	// ColorLBGRToLuv converts from LBGR to CIE Luv.
	ColorLBGRToLuv = 76

	// ColorLRGBToLuv converts from LRGB to CIE Luv.
	ColorLRGBToLuv = 77

	// ColorLabToLBGR converts from CIE Lab to LBGR.
	ColorLabToLBGR = 78

	// ColorLabToLRGB converts from CIE Lab to LRGB.
	ColorLabToLRGB = 79

	// ColorLuvToLBGR converts from CIE Luv to LBGR.
	ColorLuvToLBGR = 80

	// ColorLuvToLRGB converts from CIE Luv to LRGB.
	ColorLuvToLRGB = 81

	// ColorBGRToYUV converts from BGR to YUV.
	ColorBGRToYUV = 82

	// ColorRGBToYUV converts from RGB to YUV.
	ColorRGBToYUV = 83

	// ColorYUVToBGR converts from YUV to BGR.
	ColorYUVToBGR = 84

	// ColorYUVToRGB converts from YUV to RGB.
	ColorYUVToRGB = 85

	// ColorYUVToRGBNV12 converts from YUV 4:2:0 to RGB NV12.
	ColorYUVToRGBNV12 = 90

	// ColorYUVToBGRNV12 converts from YUV 4:2:0 to BGR NV12.
	ColorYUVToBGRNV12 = 91

	// ColorYUVToRGBNV21 converts from YUV 4:2:0 to RGB NV21.
	ColorYUVToRGBNV21 = 92

	// ColorYUVToBGRNV21 converts from YUV 4:2:0 to BGR NV21.
	ColorYUVToBGRNV21 = 93

	// ColorYUVToRGBANV12 converts from YUV 4:2:0 to RGBA NV12.
	ColorYUVToRGBANV12 = 94

	// ColorYUVToBGRANV12 converts from YUV 4:2:0 to BGRA NV12.
	ColorYUVToBGRANV12 = 95

	// ColorYUVToRGBANV21 converts from YUV 4:2:0 to RGBA NV21.
	ColorYUVToRGBANV21 = 96

	// ColorYUVToBGRANV21 converts from YUV 4:2:0 to BGRA NV21.
	ColorYUVToBGRANV21 = 97

	ColorYUVToRGBYV12 = 98
	ColorYUVToBGRYV12 = 99
	ColorYUVToRGBIYUV = 100
	ColorYUVToBGRIYUV = 101

	ColorYUVToRGBAYV12 = 102
	ColorYUVToBGRAYV12 = 103
	ColorYUVToRGBAIYUV = 104
	ColorYUVToBGRAIYUV = 105

	ColorYUVToGRAY420 = 106

	// YUV 4:2:2 family to RGB
	ColorYUVToRGBUYVY = 107
	ColorYUVToBGRUYVY = 108

	ColorYUVToRGBAUYVY = 111
	ColorYUVToBGRAUYVY = 112

	ColorYUVToRGBYUY2 = 115
	ColorYUVToBGRYUY2 = 116
	ColorYUVToRGBYVYU = 117
	ColorYUVToBGRYVYU = 118

	ColorYUVToRGBAYUY2 = 119
	ColorYUVToBGRAYUY2 = 120
	ColorYUVToRGBAYVYU = 121
	ColorYUVToBGRAYVYU = 122

	ColorYUVToGRAYUYVY = 123
	ColorYUVToGRAYYUY2 = 124

	// alpha premultiplication
	ColorRGBATomRGBA = 125
	ColormRGBAToRGBA = 126

	// RGB to YUV 4:2:0 family
	ColorRGBToYUVI420 = 127
	ColorBGRToYUVI420 = 128

	ColorRGBAToYUVI420 = 129
	ColorBGRAToYUVI420 = 130
	ColorRGBToYUVYV12  = 131
	ColorBGRToYUVYV12  = 132
	ColorRGBAToYUVYV12 = 133
	ColorBGRAToYUVYV12 = 134

	// Demosaicing
	ColorBayerBGToBGR = 46
	ColorBayerGBToBGR = 47
	ColorBayerRGToBGR = 48
	ColorBayerGRToBGR = 49

	ColorBayerBGToGRAY = 86
	ColorBayerGBToGRAY = 87
	ColorBayerRGToGRAY = 88
	ColorBayerGRToGRAY = 89

	// Demosaicing using Variable Number of Gradients
	ColorBayerBGToBGRVNG = 62
	ColorBayerGBToBGRVNG = 63
	ColorBayerRGToBGRVNG = 64
	ColorBayerGRToBGRVNG = 65

	// Edge-Aware Demosaicing
	ColorBayerBGToBGREA = 135
	ColorBayerGBToBGREA = 136
	ColorBayerRGToBGREA = 137
	ColorBayerGRToBGREA = 138

	// Demosaicing with alpha channel
	ColorBayerBGToBGRA = 139
	ColorBayerGBToBGRA = 140
	ColorBayerRGToBGRA = 141
	ColorBayerGRToBGRA = 142

	ColorCOLORCVTMAX = 143
)
