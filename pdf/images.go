// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
)

var errUnsupportedImageFormat = errors.New("unsupported image format")
var errNotJPEG = errors.New("not a JPEG")
var errBadJPEG = errors.New("bad JPEG")

type imageInfo struct {
	width            int
	height           int
	components       int
	bitsPerComponent int
}

type pdfImage struct {
	stream
	width            int
	height           int
	bitsPerComponent int
	colorSpace       string
}

func newPDFImage(seq, gen int, data []byte) *pdfImage {
	img := &pdfImage{}
	img.stream.init(seq, gen, data)
	img.dict["Type"] = name("XObject")
	img.dict["Subtype"] = name("Image")
	return img
}

func (img *pdfImage) setDimensions(width, height int) {
	img.width = width
	img.height = height
	img.dict["Width"] = integer(width)
	img.dict["Height"] = integer(height)
}

func (img *pdfImage) setBitsPerComponent(bitsPerComponent int) {
	img.bitsPerComponent = bitsPerComponent
	img.dict["BitsPerComponent"] = integer(bitsPerComponent)
}

func (img *pdfImage) setColorSpace(colorSpace string) {
	img.colorSpace = colorSpace
	img.dict["ColorSpace"] = name(colorSpace)
}

func imageColorSpace(components int) (string, error) {
	switch components {
	case 1:
		return "DeviceGray", nil
	case 3:
		return "DeviceRGB", nil
	case 4:
		return "DeviceCMYK", nil
	default:
		return "", fmt.Errorf("unsupported JPEG component count: %d", components)
	}
}

func imageKey(data []byte) string {
	sum := sha1.Sum(data)
	return fmt.Sprintf("jpeg:%x", sum)
}

func isJPEG(image []byte) bool {
	return len(image) >= 2 && image[0] == 0xFF && image[1] == 0xD8
}

func jpegInfo(image []byte) (imageInfo, error) {
	if !isJPEG(image) {
		return imageInfo{}, errNotJPEG
	}
	i := 2
	for i+3 < len(image) {
		if image[i] != 0xFF {
			return imageInfo{}, errBadJPEG
		}
		for i < len(image) && image[i] == 0xFF {
			i++
		}
		if i >= len(image) {
			return imageInfo{}, errBadJPEG
		}
		marker := image[i]
		i++
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) || marker == 0x01 {
			continue
		}
		if i+1 >= len(image) {
			return imageInfo{}, errBadJPEG
		}
		segmentLength := int(image[i])<<8 | int(image[i+1])
		i += 2
		if segmentLength < 2 || i+segmentLength-2 > len(image) {
			return imageInfo{}, errBadJPEG
		}
		switch marker {
		case 0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF:
			if segmentLength < 8 {
				return imageInfo{}, errBadJPEG
			}
			return imageInfo{
				bitsPerComponent: int(image[i]),
				height:           int(image[i+1])<<8 | int(image[i+2]),
				width:            int(image[i+3])<<8 | int(image[i+4]),
				components:       int(image[i+5]),
			}, nil
		}
		i += segmentLength - 2
	}
	return imageInfo{}, errBadJPEG
}

func readImageFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func imageInfoForData(data []byte) (imageInfo, error) {
	if isJPEG(data) {
		return jpegInfo(data)
	}
	return imageInfo{}, errUnsupportedImageFormat
}

func imageDimensions(data []byte) (width, height int, err error) {
	info, err := imageInfoForData(data)
	if err != nil {
		return 0, 0, err
	}
	return info.width, info.height, nil
}

func imageDimensionsFromFile(filename string) (width, height int, err error) {
	data, err := readImageFile(filename)
	if err != nil {
		return 0, 0, err
	}
	return imageDimensions(data)
}

func imageSizeInPoints(info imageInfo, units *units, width, height *float64) (float64, float64) {
	if width == nil && height == nil {
		return float64(info.width), float64(info.height)
	}
	if width == nil {
		h := units.toPts(*height)
		return h * float64(info.width) / float64(info.height), h
	}
	if height == nil {
		w := units.toPts(*width)
		return w, w * float64(info.height) / float64(info.width)
	}
	return units.toPts(*width), units.toPts(*height)
}

func writeImageXObject(mw *miscWriter, gw *graphWriter, name string, x, y, width, height, pageHeight float64) {
	gw.saveGraphicsState()
	gw.concatMatrix(width, 0, 0, height, x, pageHeight-y-height)
	mw.xObject(name)
	gw.restoreGraphicsState()
}

func copyImageData(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

func (dw *DocWriter) loadImage(data []byte, key string) (*pdfImage, string, error) {
	if cached, ok := dw.images[key]; ok {
		return cached.image, cached.name, nil
	}
	info, err := imageInfoForData(data)
	if err != nil {
		return nil, "", err
	}
	colorSpace, err := imageColorSpace(info.components)
	if err != nil {
		return nil, "", err
	}
	image := newPDFImage(dw.nextSeq(), 0, data)
	image.setDimensions(info.width, info.height)
	image.setBitsPerComponent(info.bitsPerComponent)
	image.setColorSpace(colorSpace)
	image.setFilter("DCTDecode")

	name := fmt.Sprintf("Im%d", len(dw.images))
	dw.file.body.add(image)
	dw.resources.setXObject(name, &indirectObjectRef{image})
	dw.images[key] = &cachedImage{image: image, name: name}
	return image, name, nil
}

func (pw *PageWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	key := imageKey(data)
	image, name, err := pw.dw.loadImage(data, key)
	if err != nil {
		return 0, 0, err
	}
	xpts := pw.units.toPts(x)
	ypts := pw.units.toPts(y)
	info := imageInfo{
		width:            image.width,
		height:           image.height,
		bitsPerComponent: image.bitsPerComponent,
	}
	wpts, hpts := imageSizeInPoints(info, pw.units, width, height)
	if pw.inPath {
		pw.endPath()
	}
	if pw.inText {
		pw.endText()
	}
	if pw.inGraph {
		pw.endGraph()
	}
	writeImageXObject(pw.mw, pw.gw, name, xpts, ypts, wpts, hpts, pw.pageHeight)
	return pw.units.fromPts(wpts), pw.units.fromPts(hpts), nil
}

func (pw *PageWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	data, err := readImageFile(filename)
	if err != nil {
		return 0, 0, err
	}
	return pw.PrintImage(data, x, y, width, height)
}
