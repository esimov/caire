package textdetector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGaussianBlur_UniformImageRemainsUniform(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 5}
	input := make([]uint8, 25)
	for i := range input {
		input[i] = 100
	}

	result, err := c.GaussianBlur(input, 1.4)

	assert.NoError(t, err)
	assert.Len(t, result, 25)
	for _, value := range result {
		assert.Equal(t, uint8(100), value)
	}
}

func TestGaussianBlur_SigmaZeroDoesNotChangeImage(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}
	input := []uint8{10, 20, 30, 40, 50, 60, 70, 80, 90}

	result, err := c.GaussianBlur(input, 0)

	assert.NoError(t, err)
	assert.Equal(t, input, result)
}

func TestGaussianBlur_NegativeSigmaDoesNotChangeImage(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}
	input := []uint8{10, 20, 30, 40, 50, 60, 70, 80, 90}

	result, err := c.GaussianBlur(input, -1)

	assert.NoError(t, err)
	assert.Equal(t, input, result)
}

func TestGaussianBlur_InvalidWidth(t *testing.T) {
	c := &CannyDetector{Width: 0, Height: 3}

	result, err := c.GaussianBlur([]uint8{1, 2, 3}, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "the pixel buffer size does not match with the image size", err.Error())
}

func TestGaussianBlur_InvalidHeight(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 0}

	result, err := c.GaussianBlur([]uint8{1, 2, 3}, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGaussianBlur_BufferTooSmall(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}

	result, err := c.GaussianBlur([]uint8{1, 2, 3, 4}, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGaussianBlur_ReturnsExpectedBufferSize(t *testing.T) {
	c := &CannyDetector{Width: 7, Height: 4}

	result, err := c.GaussianBlur(make([]uint8, 28), 1.4)

	assert.NoError(t, err)
	assert.Len(t, result, 28)
}

func TestGaussianBlur_PreservesConstantImageAtBorders(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}
	input := []uint8{200, 200, 200, 200, 200, 200, 200, 200, 200}

	result, err := c.GaussianBlur(input, 1.4)

	assert.NoError(t, err)
	for _, value := range result {
		assert.Equal(t, uint8(200), value)
	}
}
