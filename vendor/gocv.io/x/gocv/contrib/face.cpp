#include "face.h"

LBPHFaceRecognizer CreateLBPHFaceRecognizer() {
        return new cv::Ptr<cv::face::LBPHFaceRecognizer>(cv::face::LBPHFaceRecognizer::create());
}

void LBPHFaceRecognizer_Train(LBPHFaceRecognizer fr, Mats mats, IntVector labels_in) {
    std::vector<int> labels;
    for(int i = 0, *v = labels_in.val; i < labels_in.length; ++v, ++i) {
        labels.push_back(*v);
    }

    std::vector<cv::Mat> images;
    for (int i = 0; i < mats.length; ++i) {
        images.push_back(*mats.mats[i]);
    }

    (*fr)->train(images, labels);

    return;
}

void LBPHFaceRecognizer_Update(LBPHFaceRecognizer fr, Mats mats, IntVector labels_in) {
    std::vector<int> labels;
    for(int i = 0, *v = labels_in.val; i < labels_in.length; ++v, ++i) {
        labels.push_back(*v);
    }

    std::vector<cv::Mat> images;
    for (int i = 0; i < mats.length; ++i) {
        images.push_back(*mats.mats[i]);
    }

    (*fr)->update(images, labels);

    return;
}

int LBPHFaceRecognizer_Predict(LBPHFaceRecognizer fr, Mat sample) {
    int label;
    label = (*fr)->predict(*sample);

    return label;
}

struct PredictResponse LBPHFaceRecognizer_PredictExtended(LBPHFaceRecognizer fr, Mat sample) {
    struct PredictResponse response;
    int label;
    double confidence;

    (*fr)->predict(*sample, label, confidence);
     response.label = label;
     response.confidence = confidence;


    return response;
}

void LBPHFaceRecognizer_SetThreshold(LBPHFaceRecognizer fr, double threshold) {
    (*fr)->setThreshold(threshold);

    return;
}

void LBPHFaceRecognizer_SetRadius(LBPHFaceRecognizer fr, int radius) {
    (*fr)->setRadius(radius);

    return;
}

void LBPHFaceRecognizer_SetNeighbors(LBPHFaceRecognizer fr, int neighbors) {
    (*fr)->setNeighbors(neighbors);

    return;
}


int LBPHFaceRecognizer_GetNeighbors(LBPHFaceRecognizer fr) {
    int n;

    n = (*fr)->getNeighbors();

    return n;
}

void LBPHFaceRecognizer_SaveFile(LBPHFaceRecognizer fr, const char*  filename) {
    (*fr)->write(filename);

    return;
}

void LBPHFaceRecognizer_LoadFile(LBPHFaceRecognizer fr, const char*  filename) {
    (*fr)->read(filename);

    return;
}


