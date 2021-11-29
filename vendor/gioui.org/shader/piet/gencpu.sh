#!/bin/sh

# SPDX-License-Identifier: Unlicense OR MIT

set -e

OBJCOPY_ARM64=$ANDROID_SDK_ROOT/ndk/21.3.6528147/toolchains/aarch64-linux-android-4.9/prebuilt/linux-x86_64/aarch64-linux-android/bin/objcopy
OBJCOPY_ARM=$ANDROID_SDK_ROOT/ndk/21.3.6528147/toolchains/arm-linux-androideabi-4.9/prebuilt/linux-x86_64/arm-linux-androideabi/bin/objcopy
SWIFTSHADER=$HOME/.cache/swiftshader

export CGO_ENABLED=1
export GOARCH=386
export VK_ICD_FILENAMES=$SWIFTSHADER/build.32bit/Linux/vk_swiftshader_icd.json

export SWIFTSHADER_TRIPLE=armv7a-none-eabi
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer,2:image,3:image" kernel4.comp
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer" coarse.comp
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer" binning.comp
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer" backdrop.comp
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer" path_coarse.comp
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer" tile_alloc.comp
go run gioui.org/cpu/cmd/compile -arch arm -objcopy $OBJCOPY_ARM -layout "0:buffer,1:buffer,2:buffer,3:buffer" elements.comp

export GOARCH=amd64
export VK_ICD_FILENAMES=$SWIFTSHADER/build.64bit/Linux/vk_swiftshader_icd.json
export SWIFTSHADER_TRIPLE=x86_64-unknown-none-gnu

go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer,2:image,3:image" kernel4.comp
go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer" coarse.comp
go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer" binning.comp
go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer" backdrop.comp
go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer" path_coarse.comp
go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer" tile_alloc.comp
go run gioui.org/cpu/cmd/compile -arch amd64 -layout "0:buffer,1:buffer,2:buffer,3:buffer" elements.comp

export SWIFTSHADER_TRIPLE=aarch64-unknown-linux-gnu

go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer,2:image,3:image" kernel4.comp
go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer" coarse.comp
go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer" binning.comp
go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer" backdrop.comp
go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer" path_coarse.comp
go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer" tile_alloc.comp
go run gioui.org/cpu/cmd/compile -arch arm64 -objcopy $OBJCOPY_ARM64 -layout "0:buffer,1:buffer,2:buffer,3:buffer" elements.comp
