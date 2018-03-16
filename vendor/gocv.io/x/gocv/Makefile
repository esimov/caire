.ONESHELL:
.PHONY: test deps download build clean


# Package list for each well-known Linux distribution
RPMS=make cmake git gtk2-devel pkg-config libpng-devel libjpeg-devel libtiff-devel tbb tbb-devel libdc1394-devel jasper-libs jasper-devel
DEBS=unzip build-essential cmake git libgtk2.0-dev pkg-config libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libjasper-dev libdc1394-22-dev

# Detect Linux distribution
IS_FEDORA=$(shell which dnf 2>/dev/null)
IS_DEB_UBUNTU=$(shell which apt-get 2>/dev/null)
IS_RH_CENTOS=$(shell which yum 2>/dev/null)

test:
	go test .

deps:
ifneq ($(IS_FEDORA),'')
	$(MAKE) deps_fedora
else
ifneq ($(IS_RH_CENTOS),'')
	$(MAKE) deps_rh_centos
else
	$(MAKE) deps_debian
endif
endif

deps_rh_centos:
	sudo yum install $(RPMS)

deps_fedora:
	sudo dnf install $(RPMS)

deps_debian:
	sudo apt-get update
	sudo apt-get install $(DEBS)

download:
	mkdir /tmp/opencv
	cd /tmp/opencv
	wget -O opencv.zip https://github.com/opencv/opencv/archive/3.4.1.zip
	unzip opencv.zip
	wget -O opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/3.4.1.zip
	unzip opencv_contrib.zip

build:
	cd /tmp/opencv/opencv-3.4.1
	mkdir build
	cd build
	cmake -D CMAKE_BUILD_TYPE=RELEASE -D CMAKE_INSTALL_PREFIX=/usr/local -D OPENCV_EXTRA_MODULES_PATH=/tmp/opencv/opencv_contrib-3.4.1/modules -D BUILD_DOCS=OFF BUILD_EXAMPLES=OFF -D BUILD_TESTS=OFF -D BUILD_PERF_TESTS=OFF -D BUILD_opencv_java=OFF -D BUILD_opencv_python=OFF -D BUILD_opencv_python2=OFF -D BUILD_opencv_python3=OFF ..
	make -j4
	sudo make install
	sudo ldconfig

clean:
	cd ~
	rm -rf /tmp/opencv

install: download build clean
