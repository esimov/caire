#include "face_detector.h"

// FaceDetector
FaceDetector FaceDetector_New() 
{
    return new cv::Ptr<cv::pvl::FaceDetector>(cv::pvl::FaceDetector::create());
}

void FaceDetector_Close(FaceDetector f) 
{
    delete f;
}

int FaceDetector_GetRIPAngleRange(FaceDetector f) {
    return (*f)->getRIPAngleRange();
}

void FaceDetector_SetRIPAngleRange(FaceDetector f, int rip) {
    (*f)->setRIPAngleRange(rip);
}

int FaceDetector_GetROPAngleRange(FaceDetector f) {
    return (*f)->getROPAngleRange();
}

void FaceDetector_SetROPAngleRange(FaceDetector f, int rop) {
    (*f)->setROPAngleRange(rop);
}

int FaceDetector_GetMaxDetectableFaces(FaceDetector f) {
    return (*f)->getMaxDetectableFaces();
}

void FaceDetector_SetMaxDetectableFaces(FaceDetector f, int max) {
    (*f)->setMaxDetectableFaces(max);
}

int FaceDetector_GetMinFaceSize(FaceDetector f) {
    return (*f)->getMinFaceSize();
}

void FaceDetector_SetMinFaceSize(FaceDetector f, int min) {
    (*f)->setMinFaceSize(min);
}

int FaceDetector_GetBlinkThreshold(FaceDetector f) {
    return (*f)->getBlinkThreshold();
}

void FaceDetector_SetBlinkThreshold(FaceDetector f, int thresh) {
    (*f)->setBlinkThreshold(thresh);
}

int FaceDetector_GetSmileThreshold(FaceDetector f) {
    return (*f)->getSmileThreshold();
}

void FaceDetector_SetSmileThreshold(FaceDetector f, int thresh) {
    (*f)->setSmileThreshold(thresh);
}

void FaceDetector_SetTrackingModeEnabled(FaceDetector f, bool enabled)
{
    (*f)->setTrackingModeEnabled(enabled);
    return;
}

bool FaceDetector_IsTrackingModeEnabled(FaceDetector f) {
    return (*f)->isTrackingModeEnabled();
}

struct Faces FaceDetector_DetectFaceRect(FaceDetector fd, Mat img)
{
    std::vector<cv::pvl::Face> faces;
    (*fd)->detectFaceRect(*img, faces);

    Face* fs = new Face[faces.size()];
    for (size_t i = 0; i < faces.size(); ++i) {
        Face f = Face_New();
        Face_CopyTo(&faces[i], f);

        fs[i] = f;
    }
    Faces ret = {fs, (int)faces.size()};
    return ret;
}

void FaceDetector_DetectEye(FaceDetector f, Mat img, Face face)
{
    (*f)->detectEye(*img, *face);
    return;
}

void FaceDetector_DetectMouth(FaceDetector f, Mat img, Face face)
{
    (*f)->detectMouth(*img, *face);
    return;
}

void FaceDetector_DetectSmile(FaceDetector f, Mat img, Face face)
{
    (*f)->detectSmile(*img, *face);
    return;
}

void FaceDetector_DetectBlink(FaceDetector f, Mat img, Face face)
{
    (*f)->detectBlink(*img, *face);
    return;
}
