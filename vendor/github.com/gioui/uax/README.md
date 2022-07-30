<img alt="UAX Logo" src="http://npillmayer.github.io/UAX/img/UAX-Logo.svg" width="110" style="max-width:110">

### Unicode Text Segmentation Algorithms

Text processing applications need to segment text into pieces. Segments may be

* words
* sentences
* paragraphs

and so on. For western languages this is not too hard of a problem, but it may become an involved endeavor if you consider Arabic or Asian languages. From a typographic viewpoint some of these languages present serious challenges for correct segmenting. The Unicode consortium publishes recommendations and algorithms for various aspects of text segmentation in their Unicode Annexes (**UAX**).

## Text Segmentation in Go

There exist a number of Unicode standards describing best practices for text segmentation. Unfortunately, implementations in Go are sparse. Marcel van Lohuizen from the Go Core Team seems to be working on text segmenting, but with low priority. In the long run, it will be best to wait for the standard library to include functions for text segmentation. However, for now I will implement my own.

## Status

This is very much work in progress, not intended for production use. Please be patient.
