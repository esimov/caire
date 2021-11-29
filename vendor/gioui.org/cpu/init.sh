#!/bin/sh

# SPDX-License-Identifier: Unlicense OR MIT

set -e

cd ~/.cache
git clone https://github.com/eliasnaur/swiftshader
cd swiftshader

# 32-bit build
cp -a build build.32bit
cd build.32bit
CXX=clang++ CC=clang CFLAGS=-m32 CXXFLAGS=-m32 cmake -DREACTOR_EMIT_ASM_FILE=true -DSWIFTSHADER_BUILD_PVR=false -DSWIFTSHADER_BUILD_TESTS=false -DSWIFTSHADER_BUILD_GLESv2=false -DSWIFTSHADER_BUILD_EGL=false -DSWIFTSHADER_BUILD_ANGLE=false ..
cmake --build . --parallel 4
cd ..

# 64-bit build
cp -a build build.64bit
cd build.64bit
CXX=clang++ CC=clang cmake -DREACTOR_EMIT_ASM_FILE=true -DSWIFTSHADER_BUILD_PVR=false -DSWIFTSHADER_BUILD_TESTS=false -DSWIFTSHADER_BUILD_GLESv2=false -DSWIFTSHADER_BUILD_EGL=false -DSWIFTSHADER_BUILD_ANGLE=false ..
cmake --build . --parallel 4
cd ..
