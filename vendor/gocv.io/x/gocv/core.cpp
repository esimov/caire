#include "core.h"
#include <string.h>

// Mat_New creates a new empty Mat
Mat Mat_New() {
    return new cv::Mat();
}

// Mat_NewWithSize creates a new Mat with a specific size dimension and number of channels.
Mat Mat_NewWithSize(int rows, int cols, int type) {
    return new cv::Mat(rows, cols, type, 0.0);
}

// Mat_NewFromScalar creates a new Mat from a Scalar. Intended to be used
// for Mat comparison operation such as InRange.
Mat Mat_NewFromScalar(Scalar ar, int type) {
    cv::Scalar c = cv::Scalar(ar.val1, ar.val2, ar.val3, ar.val4);
    return new cv::Mat(1, 1, type, c);
}

Mat Mat_NewFromBytes(int rows, int cols, int type, struct ByteArray buf) {
    return new cv::Mat(rows, cols, type, buf.data);
}

// Mat_Close deletes an existing Mat
void Mat_Close(Mat m) {
    delete m;
}

// Mat_Empty tests if a Mat is empty
int Mat_Empty(Mat m) {
    return m->empty();
}

// Mat_Clone returns a clone of this Mat
Mat Mat_Clone(Mat m) {
    return new cv::Mat(m->clone());
}

// Mat_CopyTo copies this Mat to another Mat.
void Mat_CopyTo(Mat m, Mat dst) {
    m->copyTo(*dst);
}

// Mat_CopyToWithMask copies this Mat to another Mat while applying the mask
void Mat_CopyToWithMask(Mat m, Mat dst, Mat mask) {
    m->copyTo(*dst, *mask);
}

void Mat_ConvertTo(Mat m, Mat dst, int type) {
    m->convertTo(*dst, type);
}

// Mat_ToBytes returns the bytes representation of the underlying data.
struct ByteArray Mat_ToBytes(Mat m) {
        return toByteArray(reinterpret_cast<const char*>(m->data), m->total() * m->elemSize());
}

// Mat_Region returns a Mat of a region of another Mat
Mat Mat_Region(Mat m, Rect r) {
    return new cv::Mat(*m, cv::Rect(r.x, r.y, r.width, r.height));
}

Mat Mat_Reshape(Mat m, int cn, int rows) {
    return new cv::Mat(m->reshape(cn, rows));
}

Mat Mat_ConvertFp16(Mat m) {
    Mat dst = new cv::Mat();
    cv::convertFp16(*m, *dst);
    return dst;
}

// Mat_Mean calculates the mean value M of array elements, independently for each channel, and return it as Scalar vector
// TODO pass second paramter with mask
Scalar Mat_Mean(Mat m) {
    cv::Scalar c = cv::mean(*m);
    Scalar scal = Scalar();
    scal.val1 = c.val[0];
    scal.val2 = c.val[1];
    scal.val3 = c.val[2];
    scal.val4 = c.val[3];
    return scal;
}

void LUT(Mat src, Mat lut, Mat dst) {
  cv::LUT(*src, *lut, *dst);
}

// Mat_Rows returns how many rows in this Mat.
int Mat_Rows(Mat m) {
    return m->rows;
}

// Mat_Cols returns how many columns in this Mat.
int Mat_Cols(Mat m) {
    return m->cols;
}

// Mat_Channels returns how many channels in this Mat.
int Mat_Channels(Mat m) {
    return m->channels();
}

// Mat_Type returns the type from this Mat.
int Mat_Type(Mat m) {
    return m->type();
}

// Mat_GetUChar returns a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8U.
uint8_t Mat_GetUChar(Mat m, int row, int col) {
    return m->at<uchar>(row, col);
}

uint8_t Mat_GetUChar3(Mat m, int x, int y, int z) {
    return m->at<uchar>(x, y , z);
}

// Mat_GetSChar returns a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8S.
int8_t Mat_GetSChar(Mat m, int row, int col) {
    return m->at<schar>(row, col);
}

int8_t Mat_GetSChar3(Mat m, int x, int y, int z) {
    return m->at<schar>(x, y, z);
}

// Mat_GetShort returns a specific row/col value from this Mat expecting
// each element to contain a short aka CV_16S.
int16_t Mat_GetShort(Mat m, int row, int col) {
    return m->at<short>(row, col);
}

int16_t Mat_GetShort3(Mat m, int x, int y, int z) {
    return m->at<short>(x, y, z);
}

// Mat_GetInt returns a specific row/col value from this Mat expecting
// each element to contain an int aka CV_32S.
int32_t Mat_GetInt(Mat m, int row, int col) {
    return m->at<int>(row, col);
}

int32_t Mat_GetInt3(Mat m, int x, int y, int z) {
    return m->at<int>(x, y, z);
}

// Mat_GetFloat returns a specific row/col value from this Mat expecting
// each element to contain a float aka CV_32F.
float Mat_GetFloat(Mat m, int row, int col) {
    return m->at<float>(row, col);
}

float Mat_GetFloat3(Mat m, int x, int y, int z) {
    return m->at<float>(x, y, z);
}

// Mat_GetDouble returns a specific row/col value from this Mat expecting
// each element to contain a double aka CV_64F.
double Mat_GetDouble(Mat m, int row, int col) {
    return m->at<double>(row, col);
}

double Mat_GetDouble3(Mat m, int x, int y, int z) {
    return m->at<double>(x, y, z);
}

// Mat_SetUChar set a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8U.
void Mat_SetUChar(Mat m, int row, int col, uint8_t val) {
    m->at<uchar>(row, col) = val;
}

void Mat_SetUChar3(Mat m, int x, int y, int z, uint8_t val) {
    m->at<uchar>(x, y, z) = val;
}

// Mat_SetSChar set a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8S.
void Mat_SetSChar(Mat m, int row, int col, int8_t val) {
  m->at<schar>(row, col) = val;
}

void Mat_SetSChar3(Mat m, int x, int y, int z, int8_t val) {
  m->at<schar>(x, y, z) = val;
}

// Mat_SetShort set a specific row/col value from this Mat expecting
// each element to contain a short aka CV_16S.
void Mat_SetShort(Mat m, int row, int col, int16_t val) {
    m->at<short>(row, col) = val;
}

void Mat_SetShort3(Mat m, int x, int y, int z, int16_t val) {
    m->at<short>(x, y, z) = val;
}

// Mat_SetInt set a specific row/col value from this Mat expecting
// each element to contain an int aka CV_32S.
void Mat_SetInt(Mat m, int row, int col, int32_t val) {
    m->at<int>(row, col) = val;
}

void Mat_SetInt3(Mat m, int x, int y, int z, int32_t val) {
    m->at<int>(x, y, z) = val;
}

// Mat_SetFloat set a specific row/col value from this Mat expecting
// each element to contain a float aka CV_32F.
void Mat_SetFloat(Mat m, int row, int col, float val) {
    m->at<float>(row, col) = val;
}

void Mat_SetFloat3(Mat m, int x, int y, int z, float val) {
    m->at<float>(x, y, z) = val;
}

// Mat_SetDouble set a specific row/col value from this Mat expecting
// each element to contain a double aka CV_64F.
void Mat_SetDouble(Mat m, int row, int col, double val) {
    m->at<double>(row, col) = val;
}

void Mat_SetDouble3(Mat m, int x, int y, int z, double val) {
    m->at<double>(x, y, z) = val;
}

void Mat_AbsDiff(Mat src1, Mat src2, Mat dst) {
    cv::absdiff(*src1, *src2, *dst);
}

void Mat_Add(Mat src1, Mat src2, Mat dst) {
    cv::add(*src1, *src2, *dst);
}

void Mat_AddWeighted(Mat src1, double alpha, Mat src2, double beta, double gamma, Mat dst) {
    cv::addWeighted(*src1, alpha, *src2, beta, gamma, *dst);
}

void Mat_BitwiseAnd(Mat src1, Mat src2, Mat dst) {
    cv::bitwise_and(*src1, *src2, *dst);
}

void Mat_BitwiseNot(Mat src1, Mat dst) {
    cv::bitwise_not(*src1, *dst);
}

void Mat_BitwiseOr(Mat src1, Mat src2, Mat dst) {
    cv::bitwise_or(*src1, *src2, *dst);
}

void Mat_BitwiseXor(Mat src1, Mat src2, Mat dst) {
    cv::bitwise_xor(*src1, *src2, *dst);
}

void Mat_BatchDistance(Mat src1, Mat src2, Mat dist, int dtype, Mat nidx, int normType, int K, Mat mask, int update, bool crosscheck) {
    cv::batchDistance(*src1, *src2, *dist, dtype, *nidx, normType, K, *mask, update, crosscheck);
}

int Mat_BorderInterpolate(int p, int len, int borderType) {
    return cv::borderInterpolate(p, len, borderType);
}

void  Mat_CalcCovarMatrix(Mat samples, Mat covar, Mat mean, int flags, int ctype) {
    cv::calcCovarMatrix(*samples, *covar, *mean, flags, ctype);
}

void  Mat_CartToPolar(Mat x, Mat y, Mat magnitude, Mat angle, bool angleInDegrees) {
    cv::cartToPolar(*x, *y, *magnitude, *angle, angleInDegrees);
}

void Mat_Compare(Mat src1, Mat src2, Mat dst, int ct) {
    cv::compare(*src1, *src2, *dst, ct);
}

int Mat_CountNonZero(Mat src) {
    return cv::countNonZero(*src);
}


void Mat_CompleteSymm(Mat m, bool lowerToUpper) {
    cv::completeSymm(*m, lowerToUpper);
}

void Mat_ConvertScaleAbs(Mat src, Mat dst, double alpha, double beta) {
    cv::convertScaleAbs(*src, *dst, alpha, beta);
}

void Mat_CopyMakeBorder(Mat src, Mat dst, int top, int bottom, int left, int right, int borderType, Scalar value) {
    cv::Scalar c_value(value.val1, value.val2, value.val3, value.val4);
    cv::copyMakeBorder(*src, *dst, top, bottom, left, right, borderType, c_value);
}

void Mat_DCT(Mat src, Mat dst, int flags) {
    cv::dct(*src, *dst, flags);
}

void Mat_DFT(Mat m, Mat dst, int flags) {
    cv::dft(*m, *dst, flags);
}

void Mat_Divide(Mat src1, Mat src2, Mat dst) {
    cv::divide(*src1, *src2, *dst);
}

bool Mat_Eigen(Mat src, Mat eigenvalues, Mat eigenvectors) {
    return cv::eigen(*src, *eigenvalues, *eigenvectors);
}

void Mat_Exp(Mat src, Mat dst) {
    cv::exp(*src, *dst);
}

void Mat_ExtractChannel(Mat src, Mat dst, int coi) {
    cv::extractChannel(*src, *dst, coi);
}

void Mat_InRange(Mat src, Mat lowerb, Mat upperb, Mat dst) {
    cv::inRange(*src, *lowerb, *upperb, *dst);
}

int Mat_GetOptimalDFTSize(int vecsize) {
    return cv::getOptimalDFTSize(vecsize);
}

void Mat_Merge(struct Mats mats, Mat dst) {
    std::vector<cv::Mat> images;
    for (int i = 0; i < mats.length; ++i) {
        images.push_back(*mats.mats[i]);
    }
    cv::merge(images, *dst);
}

void Mat_MinMaxLoc(Mat m, double* minVal, double* maxVal, Point* minLoc, Point* maxLoc) {
    cv::Point cMinLoc;
    cv::Point cMaxLoc;
    cv::minMaxLoc(*m, minVal, maxVal, &cMinLoc, &cMaxLoc);

    minLoc->x = cMinLoc.x;
    minLoc->y = cMinLoc.y;
    maxLoc->x = cMaxLoc.x;
    maxLoc->y = cMaxLoc.y;
}

void Mat_Multiply(Mat src1, Mat src2, Mat dst) {
    cv::multiply(*src1, *src2, *dst);
}

void Mat_Normalize(Mat src, Mat dst, double alpha, double beta, int typ) {
    cv::normalize(*src, *dst, alpha, beta, typ);
}

double Norm(Mat src1, int normType) {
    return cv::norm(*src1, normType);
}

void Mat_Split(Mat src, struct Mats *mats) {
    std::vector<cv::Mat> channels;
    cv::split(*src, channels);
    mats->mats = new Mat[channels.size()];
    for (size_t i = 0; i < channels.size(); ++i) {
        mats->mats[i] = new cv::Mat(channels[i]);
    }
    mats->length = (int)channels.size();
}

void Mat_Subtract(Mat src1, Mat src2, Mat dst) {
    cv::subtract(*src1, *src2, *dst);
}

void Mat_Pow(Mat src, double power, Mat dst) {
    cv::pow(*src, power, *dst);
}


Scalar Mat_Sum(Mat src) {
    cv::Scalar c = cv::sum(*src);
    Scalar scal = Scalar();
    scal.val1 = c.val[0];
    scal.val2 = c.val[1];
    scal.val3 = c.val[2];
    scal.val4 = c.val[3];
    return scal;
}

// TermCriteria_New creates a new TermCriteria
TermCriteria TermCriteria_New(int typ, int maxCount, double epsilon) {
    return new cv::TermCriteria(typ, maxCount, epsilon);
}

void Contours_Close(struct Contours cs) {
    for (int i = 0; i < cs.length; i++) {
        Points_Close(cs.contours[i]);
    }
    delete[] cs.contours;
}

void KeyPoints_Close(struct KeyPoints ks) {
    delete[] ks.keypoints;
}

void Points_Close(Points ps) {
    for (size_t i = 0; i < ps.length; i++) {
        Point_Close(ps.points[i]);
    }
    delete[] ps.points;
}

void Point_Close(Point p) {}

void Rects_Close(struct Rects rs) {
    delete[] rs.rects;
}

// since it is next to impossible to iterate over mats.mats on the cgo side
Mat Mats_get(struct Mats mats, int i) {
    return mats.mats[i];
}

void Mats_Close(struct Mats mats) {
    delete[] mats.mats;
}

void ByteArray_Release(struct ByteArray buf) {
    delete[] buf.data;
}

struct ByteArray toByteArray(const char* buf, int len) {
    ByteArray ret = {new char[len], len};
    memcpy(ret.data, buf, len);
    return ret;
}
