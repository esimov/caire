uname_val="$(uname)"
if [[ "$uname_val" == "Darwin" ]]; then
  CVPATH=$(brew info opencv | grep -E "opencv/3\.([3-9]|[1-9]\d{1,})" | sed -e "s/ (.*//g")
  if [[ $CVPATH != "" ]]; then
      echo "Brew install detected"
      export CGO_CPPFLAGS="-I$CVPATH/include -I$CVPATH/include/opencv2"
      export CGO_CXXFLAGS="--std=c++1z -stdlib=libc++"
      export CGO_LDFLAGS="-L$CVPATH/lib -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d"
  else
      echo "Non-Brew install detected"
      export CGO_CPPFLAGS="-I/usr/local/include"
      export CGO_CXXFLAGS="--std=c++1z -stdlib=libc++"
      export CGO_LDFLAGS="-L/usr/local/lib -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d"
  fi

  echo "Environment variables configured for OSX"
elif [[ "$uname_val" == "Linux" ]]; then
        if [[ -f /etc/pacman.conf ]]; then
                export CGO_CPPFLAGS="-I/usr/include"
                export CGO_CXXFLAGS="--std=c++1z"
                export CGO_LDFLAGS="-L/lib64 -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d"
        else
                export PKG_CONFIG_PATH="/usr/local/lib64/pkgconfig"
                export LD_LIBRARY_PATH="/usr/local/lib64"
                export CGO_CPPFLAGS="-I/usr/local/include"
                export CGO_CXXFLAGS="--std=c++1z"
                export CGO_LDFLAGS="-L/usr/local/lib -lopencv_core -lopencv_face -lopencv_videoio -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_video -lopencv_dnn -lopencv_xfeatures2d"
        fi
  echo "Environment variables configured for Linux"
else
  echo "Unknown platform '$uname_val'!"
fi
