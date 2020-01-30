<h1 align="center"><img alt="Caire Logo" src="https://user-images.githubusercontent.com/883386/51555990-a1762600-1e81-11e9-9a6a-0cd815870358.png" height="180"></h1>

[![Build Status](https://travis-ci.org/esimov/caire.svg?branch=master)](https://travis-ci.org/esimov/caire)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/esimov/caire)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?style=flat)](./LICENSE)
[![release](https://img.shields.io/badge/release-v1.2.3-blue.svg)](https://github.com/esimov/caire/releases/tag/v1.2.3)
[![homebrew](https://img.shields.io/badge/homebrew-v1.2.3-orange.svg)](https://github.com/esimov/homebrew-caire)
[![snapcraft](https://img.shields.io/badge/snapcraft-v1.2.3-green.svg)](https://snapcraft.io/caire)

**Caire** is a content aware image resize library based on *[Seam Carving for Content-Aware Image Resizing](https://inst.eecs.berkeley.edu/~cs194-26/fa16/hw/proj4-seamcarving/imret.pdf)* paper.

### How does it work
* An energy map (edge detection) is generated from the provided image.
* The algorithm tries to find the least important parts of the image taking into account the lowest energy values.
* Using a dynamic programming approach the algorithm will generate individual seams accrossing the image from top to down, or from left to right (depending on the horizontal or vertical resizing) and will allocate for each seam a custom value, the least important pixels having the lowest energy cost and the most important ones having the highest cost.
* Traverse the image from the second row to the last row and compute the cumulative minimum energy for all possible connected seams for each entry.
* The minimum energy level is calculated by summing up the current pixel with the lowest value of the neighboring pixels from the previous row.
* Traverse the image from top to bottom and compute the minimum energy level. For each pixel in a row we compute the energy of the current pixel plus the energy of one of the three possible pixels above it.
* Find the lowest cost seam from the energy matrix starting from the last row and remove it.
* Repeat the process.

#### The process illustrated:

| Original image | Energy map | Seams applied
|:--:|:--:|:--:|
| ![original](https://user-images.githubusercontent.com/883386/35481925-de130752-0435-11e8-9246-3950679b4fd6.jpg) | ![sobel](https://user-images.githubusercontent.com/883386/35481899-5d5096ca-0435-11e8-9f9b-a84fefc06470.jpg) | ![debug](https://user-images.githubusercontent.com/883386/35481949-5c74dcb0-0436-11e8-97db-a6169cb150ca.jpg) | ![out](https://user-images.githubusercontent.com/883386/35564985-88c579d4-05c4-11e8-9068-5141714e6f43.jpg) |

## Features
Key features which differentiates this library from the other existing open source solutions:

- [x] Customizable command line support
- [x] Support for both shrinking or enlarging the image
- [x] Resize image both vertically and horizontally
- [x] Can resize all the images from a directory
- [x] Does not require any third party library
- [x] Use of sobel threshold for fine tuning
- [x] Use of blur filter for increased edge detection
- [x] Square the image with a single command
- [x] Support for proportional scaling
- [x] Face detection to avoid face deformation
- [x] Support for multiple output image type

## Face detection

The library is capable detecting human faces prior resizing the images via https://github.com/esimov/pigo, which does not require to have OpenCV installed.

The image below illustrates the application capabilities to detect human faces prior resizing. It's clearly visible from the image that with face detection activated the algorithm will avoid cropping pixels inside the detected faces, retaining the face zone unaltered.

| Original image | With face detection | Without face detection
|:--:|:--:|:--:|
| ![Original](https://user-images.githubusercontent.com/883386/37569642-0c5f49e8-2aee-11e8-8ac1-d096c0387ca0.jpg) | ![With Face Detection](https://user-images.githubusercontent.com/883386/41292871-6ca43280-6e5c-11e8-9d72-5b9a138228b6.jpg) | ![Without Face Detection](https://user-images.githubusercontent.com/883386/41292872-6cc90e8e-6e5c-11e8-8b41-5b4eb5042381.jpg) |

[Sample image source](http://www.lens-rumors.com/wp-content/uploads/2014/12/EF-M-55-200mm-f4.5-6.3-IS-STM-sample.jpg)

## Install
First, install Go, set your `GOPATH`, and make sure `$GOPATH/bin` is on your `PATH`.

```bash
$ export GOPATH="$HOME/go"
$ export PATH="$PATH:$GOPATH/bin"
```
Next download the project and build the binary file.

```bash
$ go get -u -f github.com/esimov/caire/cmd/caire
$ go install
```

## MacOS (Brew) install
The library now can be installed via Homebrew. The only thing you need is to run the commands below.

```bash
$ brew tap esimov/caire
$ brew install caire
```

## Usage

```bash
$ caire -in input.jpg -out output.jpg
```

To detect faces prior rescaling use the `-face` flag and provide the face clasification binary file included in the `data` folder. The sample code below will rescale the provided image with 20% but will check for human faces prior rescaling.

For the face detection related arguments check the Pigo [documentation](https://github.com/esimov/pigo/blob/master/README.md).

```bash
$ caire -in input.jpg -out output.jpg -face=1 -cc="data/facefinder" -perc=1 -width=20
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
| `perc` | false | Reduce image by percentage |
| `square` | false | Reduce image to square dimensions |
| `scale` | false | Proportional scaling |
| `blur` | 1 | Blur radius |
| `sobel` | 10 | Sobel filter threshold |
| `debug` | false | Use debugger |
| `face` | false | Use face detection |
| `angle` | float | Plane rotated faces angle |
| `cc` | string | Cascade classifier |

In case you wish to scale down the image by a specific percentage, it can be used the `-perc` boolean flag. In this case the values provided for the `width` and `height` options are expressed in percentage and not pixel values. For example to reduce the image dimension by 20% both horizontally and vertically you can use the following command:

```bash
$ caire -in input/source.jpg -out ./out.jpg -perc=1 -width=20 -height=20 -debug=false
```

Also the library supports the `-square` option. When this option is used the image will be resized to a squre, based on the shortest edge.

The `-scale` option will resize the image proportionally. First the image is scaled down preserving the image aspect ratio, then the seam carving algorithm is applied only to the remaining points. Ex. : given an image of dimensions 2048x1536 if we want to resize to the 1024x500, the tool first rescale the image to 1024x768, then will remove only the remaining 268px. **Using this option will drastically reduce the processing time.**

The CLI command can process all the images from a specific directory:

```bash
$ caire -in ./input-directory -out ./output-directory
```

You can also use stdin and stdout with `-`:

```bash
$ cat input/source.jpg | caire -in - -out - >out.jpg
```

`in` and `out` default to `-` so you can also use:

```bash
$ cat input/source.jpg | caire >out.jpg
```

### Caire integrations
- [x] Caire can be used as a servesless function via OpenFaaS: https://github.com/esimov/caire-openfaas
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
