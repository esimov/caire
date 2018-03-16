#!/bin/bash
set -eux -o pipefail

OPENCV_VERSION=${OPENCV_VERSION:-3.4.1}

#GRAPHICAL=ON
GRAPHICAL=${GRAPHICAL:-OFF}

# OpenCV looks for libjpeg in /usr/lib/libjpeg.so, for some reason. However,
# it does not seem to be there in 14.04. Create a link

mkdir -p $HOME/usr/lib

if [[ ! -f "$HOME/usr/lib/libjpeg.so" ]]; then
  ln -s /usr/lib/x86_64-linux-gnu/libjpeg.so $HOME/usr/lib/libjpeg.so
fi

# Same for libpng.so

if [[ ! -f "$HOME/usr/lib/libpng.so" ]]; then
  ln -s /usr/lib/x86_64-linux-gnu/libpng.so $HOME/usr/lib/libpng.so
fi

# Build OpenCV
if [[ ! -e "$HOME/usr/installed-${OPENCV_VERSION}" ]]; then
TMP=$(mktemp -d)
if [[ ! -d "opencv-${OPENCV_VERSION}/build" ]]; then
  curl -sL https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip > ${TMP}/opencv.zip
  unzip -q ${TMP}/opencv.zip
  mkdir opencv-${OPENCV_VERSION}/build
  rm ${TMP}/opencv.zip
fi

if [[ ! -d "opencv_contrib-${OPENCV_VERSION}/modules" ]]; then
   curl -sL https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip > ${TMP}/opencv-contrib.zip
   unzip -q ${TMP}/opencv-contrib.zip
   rm ${TMP}/opencv-contrib.zip
fi
rmdir ${TMP}

cd opencv-${OPENCV_VERSION}/build
cmake -D WITH_IPP=${GRAPHICAL} \
      -D WITH_OPENGL=${GRAPHICAL} \
      -D WITH_QT=${GRAPHICAL} \
      -D BUILD_EXAMPLES=OFF \
      -D BUILD_TESTS=OFF \
      -D BUILD_PERF_TESTS=OFF  \
      -D BUILD_opencv_java=OFF \
      -D BUILD_opencv_python=OFF \
      -D BUILD_opencv_python2=OFF \
      -D BUILD_opencv_python3=OFF \
      -D CMAKE_INSTALL_PREFIX=$HOME/usr \
      -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib-${OPENCV_VERSION}/modules ..
make -j8
make install && touch $HOME/usr/installed-${OPENCV_VERSION}

# caffe test data
if [[ ! -d "${HOME}/testdata" ]]; then
  mkdir ${HOME}/testdata
fi

#if [[ ! -f "${HOME}/testdata/bvlc_googlenet.prototxt" ]]; then
  cp ../../opencv-${OPENCV_VERSION}/samples/data/dnn/bvlc_googlenet.prototxt ${HOME}/testdata/bvlc_googlenet.prototxt
#fi

#if [[ ! -f "${HOME}/testdata/bvlc_googlenet.caffemodel" ]]; then
  curl -sL http://dl.caffe.berkeleyvision.org/bvlc_googlenet.caffemodel > ${HOME}/testdata/bvlc_googlenet.caffemodel
#fi

#if [[ ! -f "${HOME}/testdata/tensorflow_inception_graph.pb" ]]; then
  curl -sL https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip > ${HOME}/testdata/inception5h.zip
  unzip -o ${HOME}/testdata/inception5h.zip tensorflow_inception_graph.pb -d ${HOME}/testdata
#fi

cd ../..
touch $HOME/fresh-cache
fi
