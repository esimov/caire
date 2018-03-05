/*
Package caire is a content aware image resize library, which can rescale the source image seamlessly
both vertically and horizontally by eliminating the less important parts of the image.

The package provides a command line interface, supporting various flags for different types of rescaling operations.
To check the supported commands type:

	$ caire --help

In case you wish to integrate the API in a self constructed environment here is a simple example:

	package main

	import (
		"fmt"
		"github.com/esimov/caire"
	)

	func main() {
		p := &caire.Processor{
			// Initialize struct variables
		}

		if err := p.Process(in, out); err != nil {
			fmt.Printf("Error rescaling image: %s", err.Error())
		}
	}
 */
package caire
