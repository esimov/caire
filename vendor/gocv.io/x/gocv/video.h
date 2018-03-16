#ifndef _OPENCV3_VIDEO_H_
#define _OPENCV3_VIDEO_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
extern "C" {
#endif

#include "core.h"

#ifdef __cplusplus
typedef cv::Ptr<cv::BackgroundSubtractorMOG2>* BackgroundSubtractorMOG2;
typedef cv::Ptr<cv::BackgroundSubtractorKNN>* BackgroundSubtractorKNN;
#else
typedef void* BackgroundSubtractorMOG2;
typedef void* BackgroundSubtractorKNN;
#endif

BackgroundSubtractorMOG2 BackgroundSubtractorMOG2_Create();
void BackgroundSubtractorMOG2_Close(BackgroundSubtractorMOG2 b);
void BackgroundSubtractorMOG2_Apply(BackgroundSubtractorMOG2 b, Mat src, Mat dst);

BackgroundSubtractorKNN BackgroundSubtractorKNN_Create();
void BackgroundSubtractorKNN_Close(BackgroundSubtractorKNN b);
void BackgroundSubtractorKNN_Apply(BackgroundSubtractorKNN b, Mat src, Mat dst);

void CalcOpticalFlowPyrLK(Mat prevImg, Mat nextImg, Mat prevPts, Mat nextPts, Mat status, Mat err);
void CalcOpticalFlowFarneback(Mat prevImg, Mat nextImg, Mat flow, double pyrScale, int levels, int winsize,
	int iterations, int polyN, double polySigma, int flags);
#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_VIDEO_H_
