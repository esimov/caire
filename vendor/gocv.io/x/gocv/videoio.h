#ifndef _OPENCV3_VIDEOIO_H_
#define _OPENCV3_VIDEOIO_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
extern "C" {
#endif

#include "core.h"

#ifdef __cplusplus
typedef cv::VideoCapture* VideoCapture;
typedef cv::VideoWriter* VideoWriter;
#else
typedef void* VideoCapture;
typedef void* VideoWriter;
#endif

// VideoCapture
VideoCapture VideoCapture_New();
void VideoCapture_Close(VideoCapture v);
int VideoCapture_Open(VideoCapture v, const char* uri);
int VideoCapture_OpenDevice(VideoCapture v, int device);
void VideoCapture_Set(VideoCapture v, int prop, double param);
double VideoCapture_Get(VideoCapture v, int prop);
int VideoCapture_IsOpened(VideoCapture v);
int VideoCapture_Read(VideoCapture v, Mat buf);
void VideoCapture_Grab(VideoCapture v, int skip);

// VideoWriter
VideoWriter VideoWriter_New();
void VideoWriter_Close(VideoWriter vw);
void VideoWriter_Open(VideoWriter vw, const char* name, const char* codec, double fps, int width,
  int height);
int VideoWriter_IsOpened(VideoWriter vw);
void VideoWriter_Write(VideoWriter vw, Mat img);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_VIDEOIO_H_
