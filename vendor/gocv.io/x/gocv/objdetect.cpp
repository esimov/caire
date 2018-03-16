#include "objdetect.h"

// CascadeClassifier

CascadeClassifier CascadeClassifier_New() {
    return new cv::CascadeClassifier();
}
  
void CascadeClassifier_Close(CascadeClassifier cs) {
    delete cs;
}
  
int CascadeClassifier_Load(CascadeClassifier cs, const char* name) {
    return cs->load(name);
}
  
struct Rects CascadeClassifier_DetectMultiScale(CascadeClassifier cs, Mat img) {
    std::vector<cv::Rect> detected;
    cs->detectMultiScale(*img, detected); // uses all default parameters
    Rect* rects = new Rect[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      Rect r = {detected[i].x, detected[i].y, detected[i].width, detected[i].height};
      rects[i] = r;
    }
    Rects ret = {rects, (int)detected.size()};
    return ret;
}

struct Rects CascadeClassifier_DetectMultiScaleWithParams(CascadeClassifier cs, Mat img,
    double scale, int minNeighbors, int flags, Size minSize, Size maxSize) {

    cv::Size minSz(minSize.width, minSize.height);
    cv::Size maxSz(maxSize.width, maxSize.height);

    std::vector<cv::Rect> detected;
    cs->detectMultiScale(*img, detected, scale, minNeighbors, flags, minSz, maxSz);
    Rect* rects = new Rect[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      Rect r = {detected[i].x, detected[i].y, detected[i].width, detected[i].height};
      rects[i] = r;
    }
    Rects ret = {rects, (int)detected.size()};
    return ret;
}

// HOGDescriptor

HOGDescriptor HOGDescriptor_New() {
    return new cv::HOGDescriptor();
}

void HOGDescriptor_Close(HOGDescriptor hog) {
    delete hog;
}

int HOGDescriptor_Load(HOGDescriptor hog, const char* name) {
    return hog->load(name);
}

struct Rects HOGDescriptor_DetectMultiScale(HOGDescriptor hog, Mat img) {
    std::vector<cv::Rect> detected;
    hog->detectMultiScale(*img, detected);
    Rect* rects = new Rect[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      Rect r = {detected[i].x, detected[i].y, detected[i].width, detected[i].height};
      rects[i] = r;
    }
    Rects ret = {rects, (int)detected.size()};
    return ret;
}

struct Rects HOGDescriptor_DetectMultiScaleWithParams(HOGDescriptor hog, Mat img,
    double hitThresh, Size winStride, Size padding, double scale, double finalThresh, 
    bool useMeanshiftGrouping) {

    cv::Size wSz(winStride.width, winStride.height);
    cv::Size pSz(padding.width, padding.height);

    std::vector<cv::Rect> detected;
    hog->detectMultiScale(*img, detected, hitThresh, wSz, pSz, scale, finalThresh, useMeanshiftGrouping);
    Rect* rects = new Rect[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      Rect r = {detected[i].x, detected[i].y, detected[i].width, detected[i].height};
      rects[i] = r;
    }
    Rects ret = {rects, (int)detected.size()};
    return ret;
}

Mat HOG_GetDefaultPeopleDetector() {
    return new cv::Mat(cv::HOGDescriptor::getDefaultPeopleDetector());
}

void HOGDescriptor_SetSVMDetector(HOGDescriptor hog, Mat det) {
    hog->setSVMDetector(*det);
}

struct Rects GroupRectangles(struct Rects rects, int groupThreshold, double eps) {
    std::vector<cv::Rect> vRect;
    for (int i = 0; i < rects.length; ++i) {
        cv::Rect r = cv::Rect(rects.rects[i].x, rects.rects[i].y, rects.rects[i].width, rects.rects[i].height);
        vRect.push_back(r);
    }

    cv::groupRectangles(vRect, groupThreshold, eps);

    Rect* results = new Rect[vRect.size()];
    for (size_t i = 0; i < vRect.size(); ++i) {
      Rect r = {vRect[i].x, vRect[i].y, vRect[i].width, vRect[i].height};
      results[i] = r;
    }
    Rects ret = {results, (int)vRect.size()};
    return ret;
}
