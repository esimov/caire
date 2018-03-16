#include "face.h"

// Face
Face Face_New()
{
    return new cv::pvl::Face();
}

void Face_Close(Face face)
{
    delete face;
}

// Face_CopyTo only copies the Face Rect data.
void Face_CopyTo(Face src, Face dst)
{
    cv::Rect faceRect = src->get<cv::Rect>(cv::pvl::Face::FACE_RECT);
    int ripAngle = src->get<int>(cv::pvl::Face::RIP_ANGLE);
    int ropAngle = src->get<int>(cv::pvl::Face::ROP_ANGLE);
    int confidence = src->get<int>(cv::pvl::Face::FACE_RECT_CONFIDENCE);
    int trackingID = src->get<int>(cv::pvl::Face::TRACKING_ID);
    dst->setFaceRectInfo(faceRect, ripAngle, ropAngle, confidence, trackingID);
}

Rect Face_GetRect(Face face)
{
    cv::Rect faceRect = face->get<cv::Rect>(cv::pvl::Face::FACE_RECT);

    Rect r = {faceRect.x, faceRect.y, faceRect.width, faceRect.height};
    return r;
}

int Face_RIPAngle(Face face)
{
    return face->get<int>(cv::pvl::Face::RIP_ANGLE);
}

int Face_ROPAngle(Face face)
{
    return face->get<int>(cv::pvl::Face::ROP_ANGLE);
}

Point Face_LeftEyePosition(Face face)
{
    cv::Point pt = face->get<cv::Point>(cv::pvl::Face::LEFT_EYE_POS);
    Point p = {pt.x, pt.y};
    return p;
}

bool Face_LeftEyeClosed(Face face)
{
    return face->get<bool>(cv::pvl::Face::CLOSING_LEFT_EYE);
}

Point Face_RightEyePosition(Face face)
{
    cv::Point pt = face->get<cv::Point>(cv::pvl::Face::RIGHT_EYE_POS);
    Point p = {pt.x, pt.y};
    return p;
}

bool Face_RightEyeClosed(Face face)
{
    return face->get<bool>(cv::pvl::Face::CLOSING_RIGHT_EYE);
}

Point Face_MouthPosition(Face face)
{
    cv::Point pt = face->get<cv::Point>(cv::pvl::Face::MOUTH_POS);
    Point p = {pt.x, pt.y};
    return p;
}

bool Face_IsSmiling(Face face)
{
    return face->get<bool>(cv::pvl::Face::SMILING);
}
