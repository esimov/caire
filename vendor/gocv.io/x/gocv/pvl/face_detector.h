#ifndef _OPENCV3_FACE_DETECTOR_H_
#define _OPENCV3_FACE_DETECTOR_H_

#include <stdbool.h>

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
#include <opencv2/pvl.hpp>
extern "C" {
#endif

#include "../core.h"
#include "face.h"

#ifdef __cplusplus
typedef cv::Ptr<cv::pvl::FaceDetector>* FaceDetector;
#else
typedef void* FaceDetector;
#endif

// FaceDetector
FaceDetector FaceDetector_New();
void FaceDetector_Close(FaceDetector f);
int FaceDetector_GetRIPAngleRange(FaceDetector f);
void FaceDetector_SetRIPAngleRange(FaceDetector f, int rip);
int FaceDetector_GetROPAngleRange(FaceDetector f);
void FaceDetector_SetROPAngleRange(FaceDetector f, int rop);
int FaceDetector_GetMaxDetectableFaces(FaceDetector f);
void FaceDetector_SetMaxDetectableFaces(FaceDetector f, int max);
int FaceDetector_GetMinFaceSize(FaceDetector f);
void FaceDetector_SetMinFaceSize(FaceDetector f, int min);
int FaceDetector_GetBlinkThreshold(FaceDetector f);
void FaceDetector_SetBlinkThreshold(FaceDetector f, int thresh);
int FaceDetector_GetSmileThreshold(FaceDetector f);
void FaceDetector_SetSmileThreshold(FaceDetector f, int thresh);
void FaceDetector_SetTrackingModeEnabled(FaceDetector f, bool enabled);
bool FaceDetector_IsTrackingModeEnabled(FaceDetector f);
struct Faces FaceDetector_DetectFaceRect(FaceDetector f, Mat img);
void FaceDetector_DetectEye(FaceDetector f, Mat img, Face face);
void FaceDetector_DetectMouth(FaceDetector f, Mat img, Face face);
void FaceDetector_DetectSmile(FaceDetector f, Mat img, Face face);
void FaceDetector_DetectBlink(FaceDetector f, Mat img, Face face);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_FACE_DETECTOR_H_
