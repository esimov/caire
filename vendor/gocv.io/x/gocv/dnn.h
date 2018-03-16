#ifndef _OPENCV3_DNN_H_
#define _OPENCV3_DNN_H_

#include <stdbool.h>

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
#include <opencv2/dnn.hpp>
extern "C" {
#endif

#include "core.h"

#ifdef __cplusplus
typedef cv::dnn::Net* Net;
#else
typedef void* Net;
#endif

Net Net_ReadNetFromCaffe(const char* prototxt, const char* caffeModel);
Net Net_ReadNetFromTensorflow(const char* model);
Mat Net_BlobFromImage(Mat image, double scalefactor, Size size, Scalar mean, bool swapRB, bool crop);
void Net_Close(Net net);
bool Net_Empty(Net net);
void Net_SetInput(Net net, Mat blob, const char* name);
Mat Net_Forward(Net net, const char* outputName);

#ifdef __cplusplus
}
#endif

#endif //_OPENCV3_DNN_H_
