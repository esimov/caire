<h1 align="center"><img alt="Caire Logo" src="https://user-images.githubusercontent.com/883386/51555990-a1762600-1e81-11e9-9a6a-0cd815870358.png" height="180"></h1>

[![build](https://github.com/esimov/caire/actions/workflows/build.yml/badge.svg)](https://github.com/esimov/caire/actions/workflows/build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/esimov/caire.svg)](https://pkg.go.dev/github.com/esimov/caire)
[![license](https://img.shields.io/github/license/esimov/caire)](./LICENSE)
[![release](https://img.shields.io/badge/release-v1.4.5-blue.svg)](https://github.com/esimov/caire/releases/tag/v1.4.5)
[![homebrew](https://img.shields.io/badge/homebrew-v1.4.5-orange.svg)](https://formulae.brew.sh/formula/caire)
[![caire](https://snapcraft.io/caire/badge.svg)](https://snapcraft.io/caire)

**Caire** is a content aware image resize library based on *[Seam Carving for Content-Aware Image Resizing](https://inst.eecs.berkeley.edu/~cs194-26/fa16/hw/proj4-seamcarving/imret.pdf)* paper.

### How does it work
* An energy map (edge detection) is generated from the provided image.
* The algorithm tries to find the least important parts of the image taking into account the lowest energy values.
* Using a dynamic programming approach the algorithm will generate individual seams across the image from top to down, or from left to right (depending on the horizontal or vertical resizing) and will allocate for each seam a custom value, the least important pixels having the lowest energy cost and the most important ones having the highest cost.
* We traverse the image from the second row to the last row and compute the cumulative minimum energy for all possible connected seams for each entry.
* The minimum energy level is calculated by summing up the current pixel value with the lowest value of the neighboring pixels obtained from the previous row.
* We traverse the image from top to bottom and compute the minimum energy level. For each pixel in a row we compute the energy of the current pixel plus the energy of one of the three possible pixels above it.
* Find the lowest cost seam from the energy matrix starting from the last row and remove it.
* Repeat the process.

#### The process illustrated:

| Original image | Energy map | Seams applied
|:--:|:--:|:--:|
| ![original](https://user-images.githubusercontent.com/883386/35481925-de130752-0435-11e8-9246-3950679b4fd6.jpg) | ![sobel](https://user-images.githubusercontent.com/883386/35481899-5d5096ca-0435-11e8-9f9b-a84fefc06470.jpg) | ![debug](https://user-images.githubusercontent.com/883386/35481949-5c74dcb0-0436-11e8-97db-a6169cb150ca.jpg) | ![out](https://user-images.githubusercontent.com/883386/35564985-88c579d4-05c4-11e8-9068-5141714e6f43.jpg) |

## Features
Key features which differentiates this library from the other existing open source solutions:

- [x] **GUI progress indicator**
- [x] Customizable command line support
- [x] Support for both shrinking or enlarging the image
- [x] Resize image both vertically and horizontally
- [x] Face detection to avoid face deformation
- [x] Support for multiple output image type (jpg, jpeg, png, bmp, gif)
- [x] Support for `stdin` and `stdout` pipe commands
- [x] Can process whole directories recursively and concurrently
- [x] Use of sobel threshold for fine tuning
- [x] Use of blur filter for increased edge detection
- [x] Support for squaring the image with a single command
- [x] Support for proportional scaling
- [x] Support for protective mask
- [x] Support for removal mask
- [x] [GUI debug mode support](#masks-support)

## Install
First, install Go, set your `GOPATH`, and make sure `$GOPATH/bin` is on your `PATH`.

```bash
$ go install github.com/esimov/caire/cmd/caire@latest 
```

## MacOS (Brew) install
The library can also be installed via Homebrew.

```bash
$ brew install caire
```

## Usage

```bash
$ caire -in input.jpg -out output.jpg
```

### Supported commands:
```bash
$ caire --help
```
The following flags are supported:

| Flag | Default | Description |
| --- | --- | --- |
| `in` | - | Input file |
| `out` | - | Output file |
| `width` | n/a | New width |
| `height` | n/a | New height |
| `preview` | true | Show GUI window |
| `perc` | false | Reduce image by percentage |
| `square` | false | Reduce image to square dimensions |
| `blur` | 4 | Blur radius |
| `sobel` | 2 | Sobel filter threshold |
| `debug` | false | Use debugger |
| `face` | false | Use face detection |
| `angle` | float | Plane rotated faces angle |
| `mask` | string | Mask file path |
| `rmask` | string | Remove mask file path |
| `color` | string | Seam color (default `#ff0000`) |
| `shape` | string | Shape type used for debugging: `circle`,`line` (default `circle`) |

## Face detection

The library is capable of detecting human faces prior resizing the images by using the lightweight Pigo (https://github.com/esimov/pigo) face detection library.

The image below illustrates the application capabilities for human face detection prior resizing. It's clearly visible that with face detection activated the algorithm will avoid cropping pixels inside the detected faces, retaining the face zone unaltered.

| Original image | With face detection | Without face detection
|:--:|:--:|:--:|
| ![Original](https://user-images.githubusercontent.com/883386/37569642-0c5f49e8-2aee-11e8-8ac1-d096c0387ca0.jpg) | ![With Face Detection](https://user-images.githubusercontent.com/883386/41292871-6ca43280-6e5c-11e8-9d72-5b9a138228b6.jpg) | ![Without Face Detection](https://user-images.githubusercontent.com/883386/41292872-6cc90e8e-6e5c-11e8-8b41-5b4eb5042381.jpg) |

[Sample image source](http://www.lens-rumors.com/wp-content/uploads/2014/12/EF-M-55-200mm-f4.5-6.3-IS-STM-sample.jpg)

### GUI progress indicator

<p align="center"><img alt="GUI preview" title="GUI preview" src="https://github.com/esimov/caire/raw/master/gui_preview.gif"></p>

A GUI preview mode is also incorporated into the library for in time process visualization. The Gio GUI library has been used because of its robustness and modern architecture. Prior running it please make sure that you have installed all the required dependencies noted in the installation section (https://gioui.org/#installation) .

The preview window is activated by default but you can deactivate it any time by setting the `-preview` flag to false. When the images are processed concurrently from a directory the preview mode is deactivated.

### Face detection to avoid face deformation
In order to detect faces prior rescaling, use the `-face` flag. There is no need to provide a face classification file, since it's already embedded into the generated binary file. The sample code below will resize the provided image with 20%, but checks for human faces in order tot avoid face deformations.

For face detection related settings please check the Pigo [documentation](https://github.com/esimov/pigo/blob/master/README.md).

```bash
$ caire -in input.jpg -out output.jpg -face=1 -perc=1 -width=20
```

### Support for `stdin` and `stdout` pipe commands
You can also use `stdin` and `stdout` with `-`:

```bash
$ cat input/source.jpg | caire -in - -out - >out.jpg
```

`in` and `out` default to `-` so you can also use:

```bash
$ cat input/source.jpg | caire >out.jpg
$ caire -out out.jpg < input/source.jpg
```

You can provide also an image URL for the `-in` flag or even use **curl** or **wget** as a pipe command in which case there is no need to use the `-in` flag.

```bash
$ caire -in <image_url> -out <output-folder>
$ curl -s <image_url> | caire > out.jpg
```

### Process multiple images from a directory concurrently
The library can also process multiple images from a directory **concurrently**. You have to provide only the source and the destination folder and the new width or height in this case.

```bash
$ caire -in <input_folder> -out <output-folder>
```

### Support for multiple output image type
There is no need to define the output file type, just use the correct extension and the library will encode the image to that specific type. You can export the resized image even to a **Gif** file, in which case the generated file shows the resizing process interactively.

### Other options
In case you wish to scale down the image by a specific percentage, it can be used the **`-perc`** boolean flag. In this case the values provided for the `width` and `height` are expressed in percentage and not pixel values. For example to reduce the image dimension by 20% both horizontally and vertically you can use the following command:

```bash
$ caire -in input/source.jpg -out ./out.jpg -perc=1 -width=20 -height=20 -debug=false
```

Also the library supports the **`-square`** option. When this option is used the image will be resized to a square, based on the shortest edge.

When an image is resized on both the X and Y axis, the algorithm will first try to rescale it prior resizing, but also will preserve the image aspect ratio. The seam carving algorithm is applied only to the remaining points. Ex. : given an image of dimensions 2048x1536 if we want to resize to the 1024x500, the tool first rescale the image to 1024x768 and then will remove only the remaining 268px.

### Masks support:

- `-mask`: The path to the protective mask. The mask should be in binary format and have the same size as the input image. White areas represent regions where no seams should be carved.
- `-rmask`: The path to the removal mask. The mask should be in binary format and have the same size as the input image. White areas represent regions to be removed.

Mask | Mask removal
:-: | :-:
<video src='https://user-images.githubusercontent.com/883386/197509861-86733da8-0846-419a-95eb-4fb5a97607d5.mp4' width=180/> | <video src='https://user-images.githubusercontent.com/883386/197397857-7b785d7c-2f80-4aed-a5d2-75c429389060.mp4' width=180/>

### Caire integrations
- [x] Caire can be used as a serverless function via OpenFaaS: https://github.com/esimov/caire-openfaas
- [x] Caire can also be used as a `snap` function (https://snapcraft.io/caire): `$ snap run caire --h`

<a href="https://snapcraft.io/caire"><img src="https://raw.githubusercontent.com/snapcore/snap-store-badges/master/EN/%5BEN%5D-snap-store-white-uneditable.png" alt="snapcraft caire"></a>

## Results

#### Shrunk images
| Original | Shrunk |
| --- | --- |
| ![broadway_tower_edit](https://user-images.githubusercontent.com/883386/35498083-83d6015e-04d5-11e8-936a-883e17b76f9d.jpg) | ![broadway_tower_edit](https://user-images.githubusercontent.com/883386/35498110-a4a03328-04d5-11e8-9bf1-f526ef033d6a.jpg) |
| ![waterfall](https://user-images.githubusercontent.com/883386/35498250-2f31e202-04d6-11e8-8840-a78f40fc1a0c.png) | ![waterfall](https://user-images.githubusercontent.com/883386/35498209-0411b16a-04d6-11e8-9ce2-ec4bce34828a.jpg) |
| ![dubai](https://user-images.githubusercontent.com/883386/35498466-1375b88a-04d7-11e8-8f8e-9d202da6a6b3.jpg) | ![dubai](https://user-images.githubusercontent.com/883386/35498499-3c32fc38-04d7-11e8-9f0d-07f63a8bd420.jpg) |
| ![boat](https://user-images.githubusercontent.com/883386/35498465-1317a678-04d7-11e8-9185-ec92ea57f7c6.jpg) | ![boat](https://user-images.githubusercontent.com/883386/35498498-3c0f182c-04d7-11e8-9af8-695bc071e0f1.jpg) |

#### Enlarged images
| Original | Extended |
| --- | --- |
| ![gasadalur](https://user-images.githubusercontent.com/883386/35498662-e11853c4-04d7-11e8-98d7-fcdb27207362.jpg) | ![gasadalur](https://user-images.githubusercontent.com/883386/35498559-87eb6426-04d7-11e8-825c-2dd2abdfc112.jpg) |
| ![dubai](https://user-images.githubusercontent.com/883386/35498466-1375b88a-04d7-11e8-8f8e-9d202da6a6b3.jpg) | ![dubai](https://user-images.githubusercontent.com/883386/35498827-8cee502c-04d8-11e8-8449-05805f196d60.jpg) |
### Useful resources
* https://en.wikipedia.org/wiki/Seam_carving
* https://inst.eecs.berkeley.edu/~cs194-26/fa16/hw/proj4-seamcarving/imret.pdf
* http://pages.cs.wisc.edu/~moayad/cs766/download_files/alnammi_cs_766_final_report.pdf
* https://stacks.stanford.edu/file/druid:my512gb2187/Zargham_Nassirpour_Content_aware_image_resizing.pdf

## Author

* Endre Simo ([@simo_endre](https://twitter.com/simo_endre))

## License

Copyright © 2018 Endre Simo

This project is under the MIT License. See the LICENSE file for the full license text.
