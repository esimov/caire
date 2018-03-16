0.10.0
---
* **build** 
    * install unzip before build
    * overwrite when unzipping file to install Tensorflow test model
    * use -DCPU_DISPATCH= flag for build to avoid problem with disabled AVX on Windows
    * update unzipped file when installing Tensorflow test model
* **core** 
    * add Compare() and CountNonZero() functions
    * add getter/setter using optional params for multi-dimensional Mat using row/col/channel
    * Add mat subtract function
    * add new toRectangle function to DRY up conversion from CRects to []image.Rectangle
    * add split subtract sum wrappers
    * Add toCPoints() helper function
    * Added Mat.CopyToWithMask() per #47
    * added Pow() method
    * BatchDistance BorderInterpolate CalcCovarMatrix CartToPolar
    * CompleteSymm ConvertScaleAbs CopyMakeBorder Dct
    * divide, multiply
    * Eigen Exp ExtractChannels
    * operations on a 3d Mat are not same as a 2d multichannel Mat
    * resolve merge conflict with duplicate Subtract() function
    * run gofmt on core tests
    * Updated type for Mat.GetUCharAt() and Mat.SetUCharAt() to reflect uint8 instead of int8
* **docs** 
    * update ROADMAP of completed functions in core from recent contributions
* **env** 
    * check loading resources
    * Add distribution detection to deps rule
    * Add needed environment variables for Linux
* **highgui** 
    * add some missing test coverage on WaitKey()
* **imgproc** 
    * Add adaptive threshold function
    * Add pyrDown and pyrUp functions
    * Expose DrawContours()
    * Expose WarpPerspective and GetPerspectiveTransform
    * implement ConvexHull() and ConvexityDefects() functions
* **opencv** 
    * update to OpenCV version 3.4.1

0.9.0
---
* **bugfix** 
    * correct several errors in size parameter ordering
* **build**
    * add missing opencv_face lib reference to env.sh
    * Support for non-brew installs of opencv on Darwin
* **core**
    * add Channels() method to Mat
    * add ConvertTo() and NewMatFromBytes() functions
    * add Type() method to Mat
    * implement ConvertFp16() function
* **dnn** 
    * use correct size for blob used for Caffe/Tensorflow tests
* **docs** 
    * Update copyright date and Apache 2.0 license to include full text
* **examples** 
    * cleanup mjpeg streamer code
    * cleanup motion detector comments
    * correct use of defer in loop
    * use correct size for blob used for Caffe/Tensorflow examples
* **imgproc**
    * Add cv::approxPolyDP() bindings.
    * Add cv::arcLength() bindings.
    * Add cv::matchTemplate() bindings.
    * correct comment and link for Blur function
    * correct docs for BilateralFilter()

0.8.0
---
* **core**
    * add ColorMapFunctions and their test
    * add Mat ToBytes
    * add Reshape and MinMaxLoc functions
    * also delete points
    * fix mistake in the norm function by taking NormType instead of int as parameter
    * SetDoubleAt func and his test
    * SetFloatAt func and his test
    * SetIntAt func and his test
    * SetSCharAt func and his test
    * SetShortAt func and his test
    * SetUCharAt fun and his test
    * use correct delete operator for array of new, eliminates a bunch of memory leaks
* **dnn**
    * add support for loading Tensorflow models
    * adjust test for Caffe now that we are auto-cropping blob
    * first pass at adding Caffe support
    * go back to older function signature to avoid version conflicts with Intel CV SDK
    * properly close DNN Net class
    * use approx. value from test result to account forr windows precision differences
* **features2d**
    * implement GFTTDetector, KAZE, and MSER algorithms
    * modify MSER test for Windows results
* **highgui**
    * un-deprecate WaitKey function needed for CLI apps
* **imgcodec**
    * add fileExt type
* **imgproc**
    * add the norm wrapper and use it in test for WarpAffine and WarpAffineWithParams
    * GetRotationMatrix2D, WarpAffine and WarpAffineWithParams
    * use NormL2 in wrap affine
* **pvl**
    * add support for FaceRecognizer
    * complete wrappers for all missing FaceDetector functions
    * update instructions to match R3 of Intel CV SDK
* **docs**
    * add more detail about exactly which functions are not yet implememented in the modules that are marked as 'Work Started'
    * add refernece to Tensorflow example, and also suggest brew upgrade for MacOS
    * improve ROADMAP to help would-be contributors know where to get started
    * in the readme, explain compiling to a static library
    * remove many godoc warnings by improving function descriptions
    * update all OpenCV 3.3.1 references to v3.4.0
    * update CGO_LDFLAGS references to match latest requirements
    * update contribution guidelines to try to make it more inviting
* **examples**
    * add Caffe classifier example
    * add Tensorflow classifier example
    * fixed closing window in examples in infinite loop
    * fixed format of the examples with gofmt
* **test**
    * add helper function for test : floatEquals
    * add some attiribution from test function
    * display OpenCV version in case that test fails
    * add round function to allow for floating point accuracy differences due to GPU usage.
* **build**
    * improve search for already installed OpenCV on MacOS
    * update Appveyor build to Opencv 3.4.0
    * update to Opencv 3.4.0

0.7.0
---
* **core**
    * correct Merge implementation
* **docs**
    * change wording and formatting for roadmap
    * update roadmap for a more complete list of OpenCV functionality
    * sequence docs in README in same way as the web site, aka by OS
    * show in README that some work was done on contrib face module
* **face**
    * LBPH facerecognizer bindings
* **highgui**
    * complete implementation for remaining API functions
* **imgcodecs**
    * add IMDecode function
* **imgproc**
    * elaborate on HoughLines & HoughLinesP tests to fetch a few individual results
* **objdetect**
    * add GroupRectangles function
* **xfeatures2d**
    * add SIFT and SURF algorithms from OpenCV contrib
    * improve description for OpenCV contrib
    * run tests from OpenCV contrib

0.6.0
---
* **core**
    * Add cv::LUT binding
* **examples** 
    * do not try to go fullscreen, since does not work on OSX
* **features2d** 
    * add AKAZE algorithm
    * add BRISK algorithm
    * add FastFeatureDetector algorithm
    * implement AgastFeatureDetector algorithm
    * implement ORB algorithm
    * implement SimpleBlobDetector algorithm
* **osx**
    * Fix to get the OpenCV path with "brew info".
* **highgui** 
    * use new Window with thread lock, and deprecate WaitKey() in favor of Window.WaitKey()
    * use Window.WaitKey() in tests
* **imgproc** 
    * add tests for HoughCircles
* **pvl**
    * use correct Ptr referencing
* **video** 
    * use smart Ptr for Algorithms thanks to @alalek
    * use unsafe.Pointer for Algorithm    
    * move tests to single file now that they all pass

0.5.0
---
* **core**
    * add TermCriteria for iterative algorithms
* **imgproc**
    * add CornerSubPix() and GoodFeaturesToTrack() for corner detection
* **objdetect**
    * add DetectMultiScaleWithParams() for HOGDescriptor
    * add DetectMultiScaleWithParams() to allow override of defaults for CascadeClassifier
* **video**
    * add CalcOpticalFlowFarneback() for Farneback optical flow calculations
    * add CalcOpticalFlowPyrLK() for Lucas-Kanade optical flow calculations
* **videoio**
    * use temp directory for Windows test compat.
* **build**
    * enable Appveyor build w/cache
* **osx**
    * update env path to always match installed OpenCV from Homebrew

0.4.0
---
* **core**
    * Added cv::mean binding with single argument
    * fix the write-strings warning
    * return temp pointer fix
* **examples**
    * add counter example
    * add motion-detect command
    * correct counter
    * remove redundant cast and other small cleanup
    * set motion detect example to fullscreen
    * use MOG2 for continous motion detection, instead of simplistic first frame only
* **highgui**
    * ability to better control the fullscreen window
* **imgproc**
    * add BorderType param type for GaussianBlur
    * add BoundingRect() function
    * add ContourArea() function
    * add FindContours() function along with associated data types
    * add Laplacian and Scharr functions
    * add Moments() function
    * add Threshold function
* **pvl**
    * add needed lib for linker missing in README
* **test**
    * slightly more permissive version test
* **videoio**
    * Add image compression flags for gocv.IMWrite
    * Fixed possible looping out of compression parameters length
    * Make dedicated function to run cv::imwrite with compression parameters

0.3.1
---
* **overall**
    * Update to use OpenCV 3.3.1

0.3.0
---
* **docs** 
    * Correct Windows build location from same @jpfarias fix to gocv-site
* **core**
    * Add Resize
    * Add Mat merge and Discrete Fourier Transform
    * Add CopyTo() and Normalize()
    * Implement various important Mat logical operations
* **video**
    * BackgroundSubtractorMOG2 algorithm now working
    * Add BackgroundSubtractorKNN algorithm from video module
* **videoio**
    * Add VideoCapture::get
* **imgproc**
    * Add BilateralFilter and MedianBlur
    * Additional drawing functions implemented
    * Add HoughCircles filter
    * Implement various morphological operations
* **highgui**
    * Add Trackbar support
* **objdetect**
    * Add HOGDescriptor
* **build** 
    * Remove race from test on Travis, since it causes CGo segfault in MOG2

0.2.0
---
* Switchover to custom domain for package import
* Yes, we have Windows

0.1.0
---
Initial release!

- [X] Video capture
- [X] GUI Window to display video
- [X] Image load/save
- [X] CascadeClassifier for object detection/face tracking/etc.
- [X] Installation instructions for Ubuntu
- [X] Installation instructions for OS X
- [X] Code example to use VideoWriter
- [X] Intel CV SDK PVL FaceTracker support
- [X] imgproc Image processing
- [X] Travis CI build
- [X] At least minimal test coverage for each OpenCV class
- [X] Implement more of imgproc Image processing