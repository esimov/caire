#ifndef _OPENCV3_FACE_RECOGNIZER_H_
#define _OPENCV3_FACE_RECOGNIZER_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
#include <opencv2/pvl.hpp>
extern "C" {
#endif

#include "../core.h"
#include "face.h"

#ifdef __cplusplus
typedef cv::Ptr<cv::pvl::FaceRecognizer>* FaceRecognizer;
#else
typedef void* FaceRecognizer;
#endif

// FaceRecognizer
FaceRecognizer FaceRecognizer_New();
void FaceRecognizer_Close(FaceRecognizer f);
void FaceRecognizer_Clear(FaceRecognizer f);
bool FaceRecognizer_Empty(FaceRecognizer f);
void FaceRecognizer_SetTrackingModeEnabled(FaceRecognizer f, bool enabled);
int FaceRecognizer_CreateNewPersonID(FaceRecognizer f);
int FaceRecognizer_GetNumRegisteredPersons(FaceRecognizer f);
void FaceRecognizer_Recognize(FaceRecognizer f, Mat img, Faces faces, IntVector* pids, IntVector* confs);
int64_t FaceRecognizer_RegisterFace(FaceRecognizer f, Mat img, Face face, int personID, bool saveTofile);
void FaceRecognizer_DeregisterFace(FaceRecognizer f, int64_t faceID);
void FaceRecognizer_DeregisterPerson(FaceRecognizer f, int personID);
FaceRecognizer FaceRecognizer_Load(const char* filename);
void FaceRecognizer_Save(FaceRecognizer f, const char* filename);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_FACE_RECOGNIZER_H_
