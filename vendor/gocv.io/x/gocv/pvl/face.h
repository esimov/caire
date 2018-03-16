#ifndef _OPENCV3_FACE_H_
#define _OPENCV3_FACE_H_

#include <stdbool.h>

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
#include <opencv2/pvl.hpp>
extern "C" {
#endif

#include "../core.h"

#ifdef __cplusplus
typedef cv::pvl::Face* Face;
#else
typedef void* Face;
#endif

// Wrapper for the vector of Face struct aka std::vector<Face>
typedef struct Faces {
    Face* faces;
    int length;
} Faces;

// Face
Face Face_New();
void Face_Close(Face f);
void Face_CopyTo(Face src, Face dst);
Rect Face_GetRect(Face f);
int Face_RIPAngle(Face f);
int Face_ROPAngle(Face f);
Point Face_LeftEyePosition(Face f);
bool Face_LeftEyeClosed(Face f);
Point Face_RightEyePosition(Face f);
bool Face_RightEyeClosed(Face f);
Point Face_MouthPosition(Face f);
bool Face_IsSmiling(Face f);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_FACE_H_
