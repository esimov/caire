#ifndef _OPENCV3_FEATURES2D_H_
#define _OPENCV3_FEATURES2D_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
extern "C" {
#endif

#include "core.h"

#ifdef __cplusplus
typedef cv::Ptr<cv::AKAZE>* AKAZE;
typedef cv::Ptr<cv::AgastFeatureDetector>* AgastFeatureDetector;
typedef cv::Ptr<cv::BRISK>* BRISK;
typedef cv::Ptr<cv::FastFeatureDetector>* FastFeatureDetector;
typedef cv::Ptr<cv::GFTTDetector>* GFTTDetector;
typedef cv::Ptr<cv::KAZE>* KAZE;
typedef cv::Ptr<cv::MSER>* MSER;
typedef cv::Ptr<cv::ORB>* ORB;
typedef cv::Ptr<cv::SimpleBlobDetector>* SimpleBlobDetector;
#else
typedef void* AKAZE;
typedef void* AgastFeatureDetector;
typedef void* BRISK;
typedef void* FastFeatureDetector;
typedef void* GFTTDetector;
typedef void* KAZE;
typedef void* MSER;
typedef void* ORB;
typedef void* SimpleBlobDetector;
#endif

AKAZE AKAZE_Create();
void AKAZE_Close(AKAZE a);
struct KeyPoints AKAZE_Detect(AKAZE a, Mat src);
struct KeyPoints AKAZE_DetectAndCompute(AKAZE a, Mat src, Mat mask, Mat desc);

AgastFeatureDetector AgastFeatureDetector_Create();
void AgastFeatureDetector_Close(AgastFeatureDetector a);
struct KeyPoints AgastFeatureDetector_Detect(AgastFeatureDetector a, Mat src);

BRISK BRISK_Create();
void BRISK_Close(BRISK b);
struct KeyPoints BRISK_Detect(BRISK b, Mat src);
struct KeyPoints BRISK_DetectAndCompute(BRISK b, Mat src, Mat mask, Mat desc);

FastFeatureDetector FastFeatureDetector_Create();
void FastFeatureDetector_Close(FastFeatureDetector f);
struct KeyPoints FastFeatureDetector_Detect(FastFeatureDetector f, Mat src);

GFTTDetector GFTTDetector_Create();
void GFTTDetector_Close(GFTTDetector a);
struct KeyPoints GFTTDetector_Detect(GFTTDetector a, Mat src);

KAZE KAZE_Create();
void KAZE_Close(KAZE a);
struct KeyPoints KAZE_Detect(KAZE a, Mat src);
struct KeyPoints KAZE_DetectAndCompute(KAZE a, Mat src, Mat mask, Mat desc);

MSER MSER_Create();
void MSER_Close(MSER a);
struct KeyPoints MSER_Detect(MSER a, Mat src);

ORB ORB_Create();
void ORB_Close(ORB o);
struct KeyPoints ORB_Detect(ORB o, Mat src);
struct KeyPoints ORB_DetectAndCompute(ORB o, Mat src, Mat mask, Mat desc);

SimpleBlobDetector SimpleBlobDetector_Create();
void SimpleBlobDetector_Close(SimpleBlobDetector b);
struct KeyPoints SimpleBlobDetector_Detect(SimpleBlobDetector b, Mat src);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_FEATURES2D_H_
