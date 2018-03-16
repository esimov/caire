# Using OpenCV Contrib

The OpenCV Contrib library contains experimental or non-free (aka patented) algorithms.

GoCV support for OpenCV Contrib can be found here in the "gocv.io/x/gocv/contrib" package.

For more information about OpenCV Contrib, please go to:

https://github.com/opencv/opencv_contrib

## How to use

If you have followed the installation instructions from the main README, then the OpenCV contrib modules have already been compiled and installed.

First, you must include the `contrib` subpackage:

```go
import (
    "gocv.io/x/gocv"
    "gocv.io/x/gocv/contrib"
)
```

Then you will be able to use the functions within the `contrib` subpackage. For example, this uses the `SIFT` feature identitification algorithm that is within the `xfeatures2d` module of OpenCV:

```go
si := contrib.NewSIFT()
kp := si.Detect(img)
```
