#include "highgui.h"

// Window
void Window_New(const char* winname, int flags) {
    cv::namedWindow(winname, flags);
}

void Window_Close(const char* winname) {
    cv::destroyWindow(winname);
}

void Window_IMShow(const char* winname, Mat mat) {
    cv::imshow(winname, *mat);
}

double Window_GetProperty(const char* winname, int flag) {
    return cv::getWindowProperty(winname, flag);
}

void Window_SetProperty(const char* winname, int flag, double value) {
    cv::setWindowProperty(winname, flag, value);
}

void Window_SetTitle(const char* winname, const char* title) {
    cv::setWindowTitle(winname, title);
}

int Window_WaitKey(int delay = 0) {
    return cv::waitKey(delay);
}

void Window_Move(const char* winname, int x, int y) {
    cv::moveWindow(winname, x, y);
}

void Window_Resize(const char* winname, int width, int height) {
    cv::resizeWindow(winname, width, height);
}

struct Rect Window_SelectROI(const char* winname, Mat img) {
    cv::Rect bRect = cv::selectROI(winname, *img);
    Rect r = {bRect.x, bRect.y, bRect.width, bRect.height};
    return r;
}

struct Rects Window_SelectROIs(const char* winname, Mat img) {
    std::vector<cv::Rect> rois;
    cv::selectROIs(winname, *img, rois);
    Rect* rects = new Rect[rois.size()];
    for (size_t i = 0; i < rois.size(); ++i) {
      Rect r = {rois[i].x, rois[i].y, rois[i].width, rois[i].height};
      rects[i] = r;
    }
    Rects ret = {rects, (int)rois.size()};
    return ret;    
}

// Trackbar
void Trackbar_Create(const char* winname, const char* trackname, int max) {
    cv::createTrackbar(trackname, winname, NULL, max);
}

int Trackbar_GetPos(const char* winname, const char* trackname) {
    return cv::getTrackbarPos(trackname, winname);
}

void Trackbar_SetPos(const char* winname, const char* trackname, int pos) {
    cv::setTrackbarPos(trackname, winname, pos);
}

void Trackbar_SetMin(const char* winname, const char* trackname, int pos) {
    cv::setTrackbarMin(trackname, winname, pos);
}

void Trackbar_SetMax(const char* winname, const char* trackname, int pos) {
    cv::setTrackbarMax(trackname, winname, pos);
}
