@echo off
IF EXIST C:\opencv\build\install\include\ (
    ECHO Configuring GoCV env for OpenCV.
    set CGO_CPPFLAGS=-IC:\opencv\build\install\include
    set CGO_LDFLAGS=-LC:\opencv\build\install\x64\mingw\lib -lopencv_core341 -lopencv_face341 -lopencv_videoio341 -lopencv_imgproc341 -lopencv_highgui341 -lopencv_imgcodecs341 -lopencv_objdetect341 -lopencv_features2d341 -lopencv_video341 -lopencv_dnn341 -lopencv_xfeatures2d341
) ELSE (
    ECHO ERROR: Unable to locate OpenCV for GoCV configuration.
)
