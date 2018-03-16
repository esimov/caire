#ifndef _OPENCV3_HIGHGUI_H_
#define _OPENCV3_HIGHGUI_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
extern "C" {
#endif

#include "core.h"

// Window
void Window_New(const char* winname, int flags);
void Window_Close(const char* winname);
void Window_IMShow(const char* winname, Mat mat);
double Window_GetProperty(const char* winname, int flag);
void Window_SetProperty(const char* winname, int flag, double value);
void Window_SetTitle(const char* winname, const char* title);
int Window_WaitKey(int);
void Window_Move(const char* winname, int x, int y);
void Window_Resize(const char* winname, int width, int height);
struct Rect Window_SelectROI(const char* winname, Mat img);
struct Rects Window_SelectROIs(const char* winname, Mat img);

// Trackbar
void Trackbar_Create(const char* winname, const char* trackname, int max);
int Trackbar_GetPos(const char* winname, const char* trackname);
void Trackbar_SetPos(const char* winname, const char* trackname, int pos);
void Trackbar_SetMin(const char* winname, const char* trackname, int pos);
void Trackbar_SetMax(const char* winname, const char* trackname, int pos);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_HIGHGUI_H_
