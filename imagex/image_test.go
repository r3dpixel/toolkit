package imagex

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sunshineplan/imgconv"
)

func setupImageTest(t *testing.T) (string, func()) {
	t.Helper()
	tempImageDir, err := os.MkdirTemp("", "image_test_*")
	require.NoError(t, err, "failed to create temp dir for image tests")

	return tempImageDir, func() { _ = os.RemoveAll(tempImageDir) }
}

func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.RGBA{R: 255, A: 255}) // Red pixel
	img.Set(9, 9, color.RGBA{B: 255, A: 255}) // Blue pixel
	return img
}

func assertImageEqual(t *testing.T, expected, actual image.Image) {
	t.Helper()
	assert.Equal(t, expected.Bounds(), actual.Bounds(), "Image dimensions should match")

	keyPixels := []image.Point{{0, 0}, {9, 9}}
	for _, p := range keyPixels {
		expectedColor := expected.At(p.X, p.Y)
		actualColor := color.RGBAModel.Convert(actual.At(p.X, p.Y))
		assert.Equal(t, expectedColor, actualColor, "Pixel at %v should match", p)
	}
}

func TestFromFile(t *testing.T) {
	tempImageDir, cleanup := setupImageTest(t)
	defer cleanup()

	sourceImage := createTestImage()
	testImagePath := filepath.Join(tempImageDir, "test.png")

	t.Run("reads existing image successfully", func(t *testing.T) {
		err := ToFile(sourceImage, testImagePath, imgconv.PNG)
		require.NoError(t, err)

		readImage, err := FromFile(testImagePath)
		assert.NoError(t, err)
		require.NotNil(t, readImage)
		assertImageEqual(t, sourceImage, readImage)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempImageDir, "not-real.jpg")
		_, err := FromFile(nonExistentPath)
		assert.Error(t, err)
	})
}

func TestToFile(t *testing.T) {
	tempImageDir, cleanup := setupImageTest(t)
	defer cleanup()

	sourceImage := createTestImage()

	t.Run("writes image to file successfully", func(t *testing.T) {
		testImagePath := filepath.Join(tempImageDir, "test.png")
		err := ToFile(sourceImage, testImagePath, imgconv.PNG)
		assert.NoError(t, err)

		_, err = os.Stat(testImagePath)
		assert.NoError(t, err, "File should exist")
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		invalidPath := filepath.Join(tempImageDir, "non-existent-subdir", "test.png")
		err := ToFile(sourceImage, invalidPath, imgconv.PNG)
		assert.Error(t, err)
	})

	t.Run("writes different formats successfully", func(t *testing.T) {
		formats := []struct {
			name   string
			format imgconv.Format
			ext    string
		}{
			{"PNG", imgconv.PNG, ".png"},
			{"JPEG", imgconv.JPEG, ".jpg"},
			{"GIF", imgconv.GIF, ".gif"},
		}

		for _, tc := range formats {
			t.Run(tc.name, func(t *testing.T) {
				testPath := filepath.Join(tempImageDir, "test"+tc.ext)
				err := ToFile(sourceImage, testPath, tc.format)
				assert.NoError(t, err)

				_, err = os.Stat(testPath)
				assert.NoError(t, err)
			})
		}
	})
}

func TestFrom(t *testing.T) {
	sourceImage := createTestImage()

	t.Run("reads image from reader successfully", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := To(sourceImage, buf, imgconv.PNG)
		require.NoError(t, err)

		readImage, err := From(bytes.NewReader(buf.Bytes()))
		assert.NoError(t, err)
		require.NotNil(t, readImage)
		assertImageEqual(t, sourceImage, readImage)
	})

	t.Run("returns error for invalid data", func(t *testing.T) {
		invalidData := bytes.NewReader([]byte("not an image"))
		_, err := From(invalidData)
		assert.Error(t, err)
	})

	t.Run("returns error for empty reader", func(t *testing.T) {
		emptyReader := bytes.NewReader([]byte{})
		_, err := From(emptyReader)
		assert.Error(t, err)
	})
}

func TestTo(t *testing.T) {
	sourceImage := createTestImage()

	t.Run("writes image to writer successfully", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := To(sourceImage, buf, imgconv.PNG)
		assert.NoError(t, err)
		assert.Greater(t, buf.Len(), 0, "Buffer should contain data")

		// Verify written data is valid
		readImage, err := From(bytes.NewReader(buf.Bytes()))
		assert.NoError(t, err)
		assertImageEqual(t, sourceImage, readImage)
	})

	t.Run("writes different formats successfully", func(t *testing.T) {
		formats := []imgconv.Format{imgconv.PNG, imgconv.JPEG, imgconv.GIF}

		for _, format := range formats {
			t.Run(format.String(), func(t *testing.T) {
				buf := new(bytes.Buffer)
				err := To(sourceImage, buf, format)
				assert.NoError(t, err)
				assert.Greater(t, buf.Len(), 0)
			})
		}
	})
}

func TestFromBytes(t *testing.T) {
	sourceImage := createTestImage()

	t.Run("reads image from bytes successfully", func(t *testing.T) {
		imageBytes, err := ToBytes(sourceImage, imgconv.PNG)
		require.NoError(t, err)

		readImage, err := FromBytes(imageBytes)
		assert.NoError(t, err)
		require.NotNil(t, readImage)
		assertImageEqual(t, sourceImage, readImage)
	})

	t.Run("returns error for invalid bytes", func(t *testing.T) {
		invalidBytes := []byte("not an image")
		_, err := FromBytes(invalidBytes)
		assert.Error(t, err)
	})

	t.Run("returns error for empty bytes", func(t *testing.T) {
		_, err := FromBytes([]byte{})
		assert.Error(t, err)
	})

	t.Run("reads different formats successfully", func(t *testing.T) {
		formats := []imgconv.Format{imgconv.PNG, imgconv.JPEG, imgconv.GIF}

		for _, format := range formats {
			t.Run(format.String(), func(t *testing.T) {
				imageBytes, err := ToBytes(sourceImage, format)
				require.NoError(t, err)

				readImage, err := FromBytes(imageBytes)
				assert.NoError(t, err)
				require.NotNil(t, readImage)
			})
		}
	})
}

func TestToBytes(t *testing.T) {
	sourceImage := createTestImage()

	t.Run("converts image to bytes successfully", func(t *testing.T) {
		imageBytes, err := ToBytes(sourceImage, imgconv.PNG)
		assert.NoError(t, err)
		assert.Greater(t, len(imageBytes), 0, "Bytes should not be empty")

		// Verify bytes can be read back
		readImage, err := FromBytes(imageBytes)
		assert.NoError(t, err)
		assertImageEqual(t, sourceImage, readImage)
	})

	t.Run("converts to different formats successfully", func(t *testing.T) {
		formats := []struct {
			format imgconv.Format
			marker []byte
		}{
			{imgconv.PNG, []byte{0x89, 0x50, 0x4E, 0x47}}, // PNG magic bytes
			{imgconv.JPEG, []byte{0xFF, 0xD8, 0xFF}},      // JPEG magic bytes
			{imgconv.GIF, []byte("GIF")},                  // GIF magic bytes
		}

		for _, tc := range formats {
			t.Run(tc.format.String(), func(t *testing.T) {
				imageBytes, err := ToBytes(sourceImage, tc.format)
				assert.NoError(t, err)
				assert.Greater(t, len(imageBytes), 0)

				// Verify format by checking magic bytes
				assert.True(t, bytes.HasPrefix(imageBytes, tc.marker),
					"Should start with %s magic bytes", tc.format.String())
			})
		}
	})
}
