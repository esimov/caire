# Using the Intel Photography Vision Library

The Intel [Photography Vision Library (PVL)](https://software.intel.com/en-us/cvsdk-devguide-advanced-face-capabilities-in-intels-opencv) is a set of extensions to OpenCV that is installed with the Intel CV SDK. It uses computer vision and imaging algorithms developed at Intel.

GoCV support for the PVL can be found here in the "gocv.io/x/gocv/pvl" package.

## How to use

```go
package main

import (
	"fmt"
	"image/color"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/pvl"
)

func main() {
	deviceID := 0

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}	
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("PVL")

	// prepare input image matrix
	img := gocv.NewMat()
	defer img.Close()

	// prepare grayscale image matrix
	imgGray := gocv.NewMat()
	defer imgGray.Close()
	
	// color to draw the rect for detected faces
	blue := color.RGBA(0, 0, 255, 0)

	// load PVL FaceDetector to recognize faces
	fd := pvl.NewFaceDetector()
	defer fd.Close()

	// enable tracking mode for more efficient tracking of video source
	fd.SetTrackingModeEnabled(true)

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// convert image to grayscale for detection
		gocv.CvtColor(img, imgGray, gocv.ColorBGR2GRAY);
	
		// detect faces
		faces := fd.DetectFaceRect(imgGray)
		fmt.Printf("found %d faces\n", len(faces))

		// draw a rectangle around each face on the original image
		for _, face := range faces {
			gocv.Rectangle(img, face.Rectangle(), blue, 3)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)
	}
}
```

Some PVL examples are in the [cmd/pvl directory](../cmd/pvl) of this repo, in the form of some useful commands such as the [smile detector](../cmd/pvl/smiledetector).

## How to install the Intel CV SDK

You will need to install various dependencies before you will be able to run the Intel CV SDK installer:

```
sudo apt-get update
sudo apt-get install build-essential ffmpeg cmake checkinstall pkg-config yasm libjpeg-dev curl imagemagick gedit mplayer unzip libpng12-dev libcairo2-dev libpango1.0-dev libgtk2.0-dev libgstreamer0.10-dev libswscale.dev libavcodec-dev libavformat-dev
```

### Installing OpenCL Support

If you also want to use the OpenCL support for GPU-based hardware acceleration, you must install the OpenCL runtime. First, install the dependencies:

```
sudo apt-get update
sudo apt-get install build-essential ffmpeg cmake checkinstall pkg-config yasm libjpeg-dev curl imagemagick gedit mplayer unzip libpng12-dev libcairo2-dev libpango1.0-dev libgtk2.0-dev libgstreamer0.10-dev libswscale.dev libavcodec-dev libavformat-dev
```

Next, obtain the OpenCL runtime package:

```
wget http://registrationcenter-download.intel.com/akdlm/irc_nas/11396/SRB5.0_linux64.zip
unzip SRB5.0_linux64.zip -d SRB5.0_linux64
cd SRB5.0_linux64
```

Last, install the OpenCL runtime:

```
sudo apt-get install xz-utils
mkdir intel-opencl
tar -C intel-opencl -Jxf intel-opencl-r5.0-63503.x86_64.tar.xz
tar -C intel-opencl -Jxf intel-opencl-devel-r5.0-63503.x86_64.tar.xz
tar -C intel-opencl -Jxf intel-opencl-cpu-r5.0-63503.x86_64.tar.xz
sudo cp -R intel-opencl/* /
sudo ldconfig
```

### Installing Intel CV SDK

The most recent version of the Intel CV SDK is currently Beta R3. You can obtain it from here:

https://software.intel.com/en-us/computer-vision-sdk

One you have downloaded the compressed file, unzip the contents, and then run the `install.sh` program within the extracted directory.

## How to build/run code

Setup main Intel SDK env, by running the `setupvars.sh` program:

```
source /opt/intel/computer_vision_sdk_2017.1.163/bin/setupvars.sh
```

Then set the needed other exports for building/running GoCV code:

```
export CGO_CPPFLAGS="-I${INTEL_CVSDK_DIR}/opencv/include" CGO_LDFLAGS="-L${INTEL_CVSDK_DIR}/opencv/lib -lopencv_core -lopencv_pvl -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d"
```

You only need to do these two steps one time per session. Once you have run them, you do not need to run them again until you close your terminal window.

Now you can run the version command example to make sure you are compiling/linking against the Intel CV SDK:

```
$ go run ./cmd/version/main.go 
gocv version: 0.7.0
opencv lib version: 3.3.1-cvsdk_2017_R3.2
```

Examples that use the Intel CV SDK can be found in the `cmd/pvl` directory of this repository.
