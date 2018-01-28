# Caire

**Caire** is a content aware image resize library based on *[Seam Carving for Content-Aware Image Resizing](https://inst.eecs.berkeley.edu/~cs194-26/fa16/hw/proj4-seamcarving/imret.pdf)* paper. 

### How does it works
* An energy map (edge detection) is generated from the provided image.
* The algorithm tries to find the least important parts of the image taking into account the lowest energy values.
* Using a dynamic programming approach the algorithm will generate individual seams accrossing the image from top to down, or from left to right (depending on the horizontal or vertical resizing) and will allocate for each seam a custom value, the least important pixels having the lowest cost seam and the most important ones having the highest cost.
* Traverse the image from the second row to the last row and compute the cumulative minimum energy for all possible connected seams for each entry.
* The minimum energy level is calculated by summing up the current pixel value with the lowest value of the neighboring pixels from the previous row.
* Traverse the image from top to bottom and compute the minimum energy level. For each pixel in a row we compute the energy of the current pixel plus the energy of one of the three possible pixels above it.
* Find the lowest cost seam from the energy matrix starting from the last row and remove it.
* Repeat

#### The process illustrated:

| Original image | Energy map | Seams applied
|:--:|:--:|:--:|
| ![original](https://user-images.githubusercontent.com/883386/35481925-de130752-0435-11e8-9246-3950679b4fd6.jpg) | ![sobel](https://user-images.githubusercontent.com/883386/35481899-5d5096ca-0435-11e8-9f9b-a84fefc06470.jpg) | ![out](https://user-images.githubusercontent.com/883386/35481949-5c74dcb0-0436-11e8-97db-a6169cb150ca.jpg) || 

## Features
Key features which differentiates from the other existing open source solutions:

- [x] Very customizable
- [x] Supports for image decrease and increase as well
- [x] A very customizable command line support
- [x] Does not require any third party library
- [x] Use of sobel threshold for fine tuning

### To Do
- [ ] Face detection

## Usage

```bash
$ caire -in input.jpg -out output.jpg
```

### Supported commands:
```bash 
$ caire --help

  -in string
    	Source
  -out string
    	Destination  
  -width int
    	New width
  -height int
    	New height
  -perc
    	Reduce image by percentage
  -blur int
    	Blur radius (default 1)
  -debug
    	Use debugger
  -sobel int
    	Sobel filter threshold (default 10)
```

### Useful resources
* https://en.wikipedia.org/wiki/Seam_carving
* https://inst.eecs.berkeley.edu/~cs194-26/fa16/hw/proj4-seamcarving/imret.pdf
* http://pages.cs.wisc.edu/~moayad/cs766/download_files/alnammi_cs_766_final_report.pdf
* https://stacks.stanford.edu/file/druid:my512gb2187/Zargham_Nassirpour_Content_aware_image_resizing.pdf

