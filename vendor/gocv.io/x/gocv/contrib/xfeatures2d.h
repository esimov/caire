#ifndef _OPENCV3_XFEATURES2D_H_
#define _OPENCV3_XFEATURES2D_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
#include <opencv2/xfeatures2d.hpp>
extern "C" {
#endif

#include "../core.h"

#ifdef __cplusplus
typedef cv::Ptr<cv::xfeatures2d::SIFT>* SIFT;
typedef cv::Ptr<cv::xfeatures2d::SURF>* SURF;
#else
typedef void* SIFT;
typedef void* SURF;
#endif

SIFT SIFT_Create();
void SIFT_Close(SIFT f);
struct KeyPoints SIFT_Detect(SIFT f, Mat src);
struct KeyPoints SIFT_DetectAndCompute(SIFT f, Mat src, Mat mask, Mat desc);

SURF SURF_Create();
void SURF_Close(SURF f);
struct KeyPoints SURF_Detect(SURF f, Mat src);
struct KeyPoints SURF_DetectAndCompute(SURF f, Mat src, Mat mask, Mat desc);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_XFEATURES2D_H_
