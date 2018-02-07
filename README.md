# Caire

[![Build Status](https://travis-ci.org/esimov/caire.svg?branch=master)](https://travis-ci.org/esimov/caire)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?style=flat)](./LICENSE)
[![homebrew](https://img.shields.io/badge/homebrew-v1.0.1-orange.svg)]()

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
Key features which differentiates from the other existing open source solutions:

- [x] Customizable command line support
- [x] Support for both shrinking or enlarging the image
- [x] Resize image both vertically and horizontally
- [x] Can resize all the images from a directory
- [x] Does not require any third party library
- [x] Use of sobel threshold for fine tuning
- [x] Use of blur filter for increased edge detection

### To Do
- [ ] Face detection

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

### Supported commands:
```bash 
$ caire --help
```
The following flags are supported:

| Flag | Default | Description |
| --- | --- | --- |
| `in` | n/a | Input file |
| `out` | n/a | Output file |
| `width` | n/a | New width |
| `height` | n/a | New height |
| `perc` | false | Reduce image by percentage |
| `blur` | 1 | Blur radius |
| `sobel` | 10 | Sobel filter threshold |
| `debug` | false | Use debugger |

In case you wish to reduce the image size by a specific percentage, it can be used the `-perc` boolean flag, which means the image will be reduced to the width and height expressed as percentage. Here is a sample command using `-perc`:

```bash
caire -in input/source.jpg -out ./out.jpg -perc=1 -width=20 -height=20 -debug=false
```

which reduces the image width and height by 20%.

The CLI command can process all the images from a specific directory too.

```bash
$ caire -in ./input-directory -out ./output-directory
```

## Sample images

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


## License
This project is under the MIT License. See the LICENSE file for the full license text.
