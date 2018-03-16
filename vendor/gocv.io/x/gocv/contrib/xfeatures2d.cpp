#include "xfeatures2d.h"

SIFT SIFT_Create() {
    // TODO: params
    return new cv::Ptr<cv::xfeatures2d::SIFT>(cv::xfeatures2d::SIFT::create());
}

void SIFT_Close(SIFT d) {
    delete d;
}

struct KeyPoints SIFT_Detect(SIFT d, Mat src) {
    std::vector<cv::KeyPoint> detected;
    (*d)->detect(*src, detected);

    KeyPoint* kps = new KeyPoint[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      KeyPoint k = {detected[i].pt.x, detected[i].pt.y, detected[i].size, detected[i].angle,
        detected[i].response, detected[i].octave, detected[i].class_id};
      kps[i] = k;
    }
    KeyPoints ret = {kps, (int)detected.size()};
    return ret;
}

struct KeyPoints SIFT_DetectAndCompute(SIFT d, Mat src, Mat mask, Mat desc) {
    std::vector<cv::KeyPoint> detected;
    (*d)->detectAndCompute(*src, *mask, detected, *desc);

    KeyPoint* kps = new KeyPoint[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      KeyPoint k = {detected[i].pt.x, detected[i].pt.y, detected[i].size, detected[i].angle,
        detected[i].response, detected[i].octave, detected[i].class_id};
      kps[i] = k;
    }
    KeyPoints ret = {kps, (int)detected.size()};
    return ret;
}

SURF SURF_Create() {
    // TODO: params
    return new cv::Ptr<cv::xfeatures2d::SURF>(cv::xfeatures2d::SURF::create());
}

void SURF_Close(SURF d) {
    delete d;
}

struct KeyPoints SURF_Detect(SURF d, Mat src) {
    std::vector<cv::KeyPoint> detected;
    (*d)->detect(*src, detected);

    KeyPoint* kps = new KeyPoint[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      KeyPoint k = {detected[i].pt.x, detected[i].pt.y, detected[i].size, detected[i].angle,
        detected[i].response, detected[i].octave, detected[i].class_id};
      kps[i] = k;
    }
    KeyPoints ret = {kps, (int)detected.size()};
    return ret;
}

struct KeyPoints SURF_DetectAndCompute(SURF d, Mat src, Mat mask, Mat desc) {
    std::vector<cv::KeyPoint> detected;
    (*d)->detectAndCompute(*src, *mask, detected, *desc);

    KeyPoint* kps = new KeyPoint[detected.size()];
    for (size_t i = 0; i < detected.size(); ++i) {
      KeyPoint k = {detected[i].pt.x, detected[i].pt.y, detected[i].size, detected[i].angle,
        detected[i].response, detected[i].octave, detected[i].class_id};
      kps[i] = k;
    }
    KeyPoints ret = {kps, (int)detected.size()};
    return ret;
}
