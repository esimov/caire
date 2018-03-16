#ifndef _OPENCV3_CORE_H_
#define _OPENCV3_CORE_H_

#include <stdint.h>
#include <stdbool.h>

typedef struct String {
  const char* str;
  int length;
} String;

typedef struct ByteArray{
  char *data;
  int length;
} ByteArray;

// Wrapper for std::vector<int>
typedef struct IntVector {
  int *val;
  int length;
} IntVector;

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
extern "C" {
#endif

typedef struct RawData {
  int width;
  int height;
  struct ByteArray data;
} RawData;

// Wrapper for an individual cv::cvPoint
typedef struct Point {
  int x;
  int y;
} Point;

// Wrapper for the vector of Point structs aka std::vector<Point>
typedef struct Points {
  Point* points;
  int length;
} Points;

// Contour is alias for Points
typedef Points Contour;

// Wrapper for the vector of Points vectors aka std::vector< std::vector<Point> >
typedef struct Contours {
  Contour* contours;
  int length;
} Contours;

// Wrapper for an individual cv::cvRect
typedef struct Rect {
  int x;
  int y;
  int width;
  int height;
} Rect;

// Wrapper for the vector of Rect struct aka std::vector<Rect>
typedef struct Rects {
  Rect* rects;
  int length;
} Rects;

// Wrapper for an individual cv::cvSize
typedef struct Size {
  int width;
  int height;
} Size;

// Wrapper for an individual cv::cvScalar
typedef struct Scalar {
  double val1;
  double val2;
  double val3;
  double val4;
} Scalar;

// Wrapper for a individual cv::KeyPoint
typedef struct KeyPoint {
  double x;
  double y;
  double size;
  double angle;
  double response;
  int octave;
  int classID;
} KeyPoint;

// Wrapper for the vector of KeyPoint struct aka std::vector<KeyPoint>
typedef struct KeyPoints {
  KeyPoint* keypoints;
  int length;
} KeyPoints;

// Wrapper for an individual cv::Moment
typedef struct Moment {
  double m00;
  double m10;
  double m01;
  double m20;
  double m11;
  double m02;
  double m30;
  double m21;
  double m12;
  double m03;

  double mu20;
  double mu11;
  double mu02;
  double mu30;
  double mu21;
  double mu12;
  double mu03;

  double nu20;
  double nu11;
  double nu02;
  double nu30;
  double nu21;
  double nu12;
  double nu03;
} Moment;

#ifdef __cplusplus
typedef cv::Mat* Mat;
typedef cv::TermCriteria* TermCriteria;
#else
typedef void* Mat;
typedef void* TermCriteria;
#endif

// Wrapper for the vector of Mat aka std::vector<Mat>
typedef struct Mats {
  Mat* mats;
  int length;
} Mats;

Mat Mats_get(struct Mats mats, int i);

struct ByteArray toByteArray(const char* buf, int len);
void ByteArray_Release(struct ByteArray buf);

void Contours_Close(struct Contours cs);
void KeyPoints_Close(struct KeyPoints ks);
void Rects_Close(struct Rects rs);
void Mats_Close(struct Mats mats);
void Point_Close(struct Point p);
void Points_Close(struct Points ps);

Mat Mat_New();
Mat Mat_NewWithSize(int rows, int cols, int type);
Mat Mat_NewFromScalar(const Scalar ar, int type);
Mat Mat_NewFromBytes(int rows, int cols, int type, struct ByteArray buf);
void Mat_Close(Mat m);
int Mat_Empty(Mat m);
Mat Mat_Clone(Mat m);
void Mat_CopyTo(Mat m, Mat dst);
void Mat_CopyToWithMask(Mat m, Mat dst, Mat mask);
void Mat_ConvertTo(Mat m, Mat dst, int type);
struct ByteArray Mat_ToBytes(Mat m);
Mat Mat_Region(Mat m, Rect r);
Mat Mat_Reshape(Mat m, int cn, int rows);
Mat Mat_ConvertFp16(Mat m);
Scalar Mat_Mean(Mat m);
int Mat_Rows(Mat m);
int Mat_Cols(Mat m);
int Mat_Channels(Mat m);
int Mat_Type(Mat m);

uint8_t Mat_GetUChar(Mat m, int row, int col);
uint8_t Mat_GetUChar3(Mat m, int x, int y, int z);
int8_t Mat_GetSChar(Mat m, int row, int col);
int8_t Mat_GetSChar3(Mat m, int x, int y, int z);
int16_t Mat_GetShort(Mat m, int row, int col);
int16_t Mat_GetShort3(Mat m, int x, int y, int z);
int32_t Mat_GetInt(Mat m, int row, int col);
int32_t Mat_GetInt3(Mat m, int x, int y, int z);
float Mat_GetFloat(Mat m, int row, int col);
float Mat_GetFloat3(Mat m, int x, int y, int z);
double Mat_GetDouble(Mat m, int row, int col);
double Mat_GetDouble3(Mat m, int x, int y, int z);

void Mat_SetUChar(Mat m, int row, int col, uint8_t val);
void Mat_SetUChar3(Mat m, int x, int y, int z, uint8_t val);
void Mat_SetSChar(Mat m, int row, int col, int8_t val);
void Mat_SetSChar3(Mat m, int x, int y, int z, int8_t val);
void Mat_SetShort(Mat m, int row, int col, int16_t val);
void Mat_SetShort3(Mat m, int x, int y, int z, int16_t val);
void Mat_SetInt(Mat m, int row, int col, int32_t val);
void Mat_SetInt3(Mat m, int x, int y, int z, int32_t val);
void Mat_SetFloat(Mat m, int row, int col, float val);
void Mat_SetFloat3(Mat m, int x, int y, int z, float val);
void Mat_SetDouble(Mat m, int row, int col, double val);
void Mat_SetDouble3(Mat m, int x, int y, int z, double val);

void LUT(Mat src, Mat lut, Mat dst);

void Mat_AbsDiff(Mat src1, Mat src2, Mat dst);
void Mat_Add(Mat src1, Mat src2, Mat dst);
void Mat_AddWeighted(Mat src1, double alpha, Mat src2, double beta, double gamma, Mat dst);
void Mat_BitwiseAnd(Mat src1, Mat src2, Mat dst);
void Mat_BitwiseNot(Mat src1, Mat dst);
void Mat_BitwiseOr(Mat src1, Mat src2, Mat dst);
void Mat_BitwiseXor(Mat src1, Mat src2, Mat dst);
void Mat_Compare(Mat src1, Mat src2, Mat dst, int ct);
void Mat_BatchDistance(Mat src1, Mat src2, Mat dist, int dtype, Mat nidx, int normType, int K, Mat mask, int update, bool crosscheck);
int Mat_BorderInterpolate(int p, int len, int borderType);
void Mat_CalcCovarMatrix(Mat samples, Mat covar, Mat mean, int flags, int ctype);
void Mat_CartToPolar(Mat x, Mat y, Mat magnitude, Mat angle, bool angleInDegrees);
void Mat_CompleteSymm(Mat m, bool lowerToUpper);
void Mat_ConvertScaleAbs(Mat src, Mat dst, double alpha, double beta);
void Mat_CopyMakeBorder(Mat src, Mat dst, int top, int bottom, int left, int right, int borderType, Scalar value);
int Mat_CountNonZero(Mat src);
void Mat_DCT(Mat src, Mat dst, int flags);
void Mat_DFT(Mat m, Mat dst, int flags);
void Mat_Divide(Mat src1, Mat src2, Mat dst);
bool Mat_Eigen(Mat src, Mat eigenvalues, Mat eigenvectors);
void Mat_Exp(Mat src, Mat dst);
void Mat_InRange(Mat src, Mat lowerb, Mat upperb, Mat dst);
int Mat_GetOptimalDFTSize(int vecsize);
void Mat_ExtractChannel(Mat src, Mat dst, int coi);
void Mat_Merge(struct Mats mats, Mat dst);
void Mat_MinMaxLoc(Mat m, double* minVal, double* maxVal, Point* minLoc, Point* maxLoc);
void Mat_Multiply(Mat src1, Mat src2, Mat dst);
void Mat_Subtract(Mat src1, Mat src2, Mat dst);
void Mat_Normalize(Mat src, Mat dst, double alpha, double beta, int typ);
double Norm(Mat src1, int normType);
void Mat_Split(Mat src, struct Mats *mats);
void Mat_Subtract(Mat src1, Mat src2, Mat dst);
void Mat_Pow(Mat src, double power, Mat dst);
Scalar Mat_Sum(Mat src1);

TermCriteria TermCriteria_New(int typ, int maxCount, double epsilon);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_CORE_H_
