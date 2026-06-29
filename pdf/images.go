// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"io"

	"github.com/rowland/leadtype/svg"
)

var errUnsupportedImageFormat = errors.New("unsupported image format")
var errNotJPEG = errors.New("not a JPEG")
var errBadJPEG = errors.New("bad JPEG")
var errNotPNG = errors.New("not a PNG")
var errBadPNG = errors.New("bad PNG")
var errUnsupportedPNG = errors.New("unsupported PNG")

type imageInfo struct {
	width            int
	height           int
	components       int
	bitsPerComponent int
}

const (
	imageComponentsGray      = 1
	imageComponentsGrayAlpha = 2
	imageComponentsRGB       = 3
	imageComponentsRGBA      = 4
	imageComponentsCMYK      = 4
)

const (
	pngColorTypeGray      = 0
	pngColorTypeRGB       = 2
	pngColorTypePalette   = 3
	pngColorTypeGrayAlpha = 4
	pngColorTypeRGBA      = 6
)

type pdfImage struct {
	stream
	width            int
	height           int
	bitsPerComponent int
	colorSpace       string
}

type decodedImage struct {
	info      imageInfo
	data      []byte
	filter    string
	alphaData []byte
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

func (img *pdfImage) setSMask(ref *indirectObjectRef) {
	img.dict["SMask"] = ref
}

func imageColorSpace(components int) (string, error) {
	switch components {
	case imageComponentsGray:
		return "DeviceGray", nil
	case imageComponentsRGB:
		return "DeviceRGB", nil
	case imageComponentsCMYK:
		return "DeviceCMYK", nil
	default:
		return "", fmt.Errorf("unsupported JPEG component count: %d", components)
	}
}

func imageKey(data []byte) string {
	sum := sha1.Sum(data)
	switch {
	case svg.LooksLikeSVG(data):
		return fmt.Sprintf("svg:%x", sum)
	case isJPEG(data):
		return fmt.Sprintf("jpeg:%x", sum)
	case isPNG(data):
		return fmt.Sprintf("png:%x", sum)
	default:
		return fmt.Sprintf("image:%x", sum)
	}
}

func isJPEG(image []byte) bool {
	return len(image) >= 2 && image[0] == 0xFF && image[1] == 0xD8
}

func isPNG(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	sig := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	for i := range sig {
		if data[i] != sig[i] {
			return false
		}
	}
	return true
}

func jpegInfo(image []byte) (imageInfo, error) {
	if !isJPEG(image) {
		return imageInfo{}, errNotJPEG
	}
	return jpegInfoFromReader(bytes.NewReader(image))
}

func jpegInfoFromReader(r io.Reader) (imageInfo, error) {
	var soi [2]byte
	if _, err := io.ReadFull(r, soi[:]); err != nil {
		return imageInfo{}, errBadJPEG
	}
	if soi[0] != 0xFF || soi[1] != 0xD8 {
		return imageInfo{}, errNotJPEG
	}
	for {
		var b [1]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return imageInfo{}, errBadJPEG
		}
		if b[0] != 0xFF {
			return imageInfo{}, errBadJPEG
		}
		for b[0] == 0xFF {
			if _, err := io.ReadFull(r, b[:]); err != nil {
				return imageInfo{}, errBadJPEG
			}
		}
		marker := b[0]
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) || marker == 0x01 {
			continue
		}
		var lengthBytes [2]byte
		if _, err := io.ReadFull(r, lengthBytes[:]); err != nil {
			return imageInfo{}, errBadJPEG
		}
		segmentLength := int(lengthBytes[0])<<8 | int(lengthBytes[1])
		if segmentLength < 2 {
			return imageInfo{}, errBadJPEG
		}
		switch marker {
		case 0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF:
			if segmentLength < 8 {
				return imageInfo{}, errBadJPEG
			}
			var sof [6]byte
			if _, err := io.ReadFull(r, sof[:]); err != nil {
				return imageInfo{}, errBadJPEG
			}
			return imageInfo{
				bitsPerComponent: int(sof[0]),
				height:           int(sof[1])<<8 | int(sof[2]),
				width:            int(sof[3])<<8 | int(sof[4]),
				components:       int(sof[5]),
			}, nil
		default:
			if _, err := io.CopyN(io.Discard, r, int64(segmentLength-2)); err != nil {
				return imageInfo{}, errBadJPEG
			}
		}
	}
}

func pngInfo(data []byte) (imageInfo, error) {
	if !isPNG(data) {
		return imageInfo{}, errNotPNG
	}
	if len(data) < 33 {
		return imageInfo{}, errBadPNG
	}
	if string(data[12:16]) != "IHDR" {
		return imageInfo{}, errBadPNG
	}
	if binary.BigEndian.Uint32(data[8:12]) != 13 {
		return imageInfo{}, errBadPNG
	}
	width := int(binary.BigEndian.Uint32(data[16:20]))
	height := int(binary.BigEndian.Uint32(data[20:24]))
	if width <= 0 || height <= 0 {
		return imageInfo{}, errBadPNG
	}
	bitDepth := int(data[24])
	colorType := int(data[25])
	compressionMethod := data[26]
	filterMethod := data[27]
	interlaceMethod := data[28]
	if compressionMethod != 0 || filterMethod != 0 {
		return imageInfo{}, errUnsupportedPNG
	}
	if interlaceMethod != 0 {
		return imageInfo{}, errUnsupportedPNG
	}
	components := 0
	switch colorType {
	case pngColorTypeGray:
		if bitDepth != 1 && bitDepth != 2 && bitDepth != 4 && bitDepth != 8 {
			return imageInfo{}, errUnsupportedPNG
		}
		components = imageComponentsGray
	case pngColorTypeRGB:
		if bitDepth != 8 {
			return imageInfo{}, errUnsupportedPNG
		}
		components = imageComponentsRGB
	case pngColorTypePalette:
		if bitDepth != 1 && bitDepth != 2 && bitDepth != 4 && bitDepth != 8 {
			return imageInfo{}, errUnsupportedPNG
		}
		components = imageComponentsRGBA
	case pngColorTypeGrayAlpha:
		if bitDepth != 8 {
			return imageInfo{}, errUnsupportedPNG
		}
		components = imageComponentsGrayAlpha
	case pngColorTypeRGBA:
		if bitDepth != 8 {
			return imageInfo{}, errUnsupportedPNG
		}
		components = imageComponentsRGBA
	default:
		return imageInfo{}, errUnsupportedPNG
	}
	return imageInfo{
		width:            width,
		height:           height,
		components:       components,
		bitsPerComponent: 8,
	}, nil
}

func pngInfoFromReader(r io.Reader) (imageInfo, error) {
	header := make([]byte, 33)
	if _, err := io.ReadFull(r, header); err != nil {
		return imageInfo{}, errBadPNG
	}
	return pngInfo(header)
}

func imageInfoForData(data []byte) (imageInfo, error) {
	if svg.LooksLikeSVG(data) {
		width, height, err := svgDimensions(data)
		if err != nil {
			return imageInfo{}, err
		}
		return imageInfo{width: width, height: height}, nil
	}
	if isJPEG(data) {
		return jpegInfo(data)
	}
	if isPNG(data) {
		return pngInfo(data)
	}
	return imageInfo{}, errUnsupportedImageFormat
}

func imageInfoFromReader(r io.Reader) (imageInfo, error) {
	br := bufio.NewReader(r)
	header, _ := br.Peek(8)
	if len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8 {
		return jpegInfoFromReader(br)
	}
	if isPNG(header) {
		return pngInfoFromReader(br)
	}
	header, _ = br.Peek(512)
	if svg.LooksLikeSVG(header) {
		width, height, err := svgDimensionsFromReader(br)
		if err != nil {
			return imageInfo{}, err
		}
		return imageInfo{width: width, height: height}, nil
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

func writeImageXObject(mw *miscWriter, gw *graphWriter, name string, x, y, width, height, pageHeight float64, extGStateName string) {
	gw.saveGraphicsState()
	if extGStateName != "" {
		mw.setExtGState(extGStateName)
	}
	gw.concatMatrix(width, 0, 0, height, x, pageHeight-y-height)
	mw.xObject(name)
	gw.restoreGraphicsState()
}

func writeFormXObject(mw *miscWriter, gw *graphWriter, name string, x, y, width, height, sourceWidth, sourceHeight, pageHeight float64, extGStateName string) {
	gw.saveGraphicsState()
	if extGStateName != "" {
		mw.setExtGState(extGStateName)
	}
	gw.concatMatrix(width/sourceWidth, 0, 0, height/sourceHeight, x, pageHeight-y-height)
	mw.xObject(name)
	gw.restoreGraphicsState()
}

func decodePNG(data []byte) (decodedImage, error) {
	info, err := pngInfo(data)
	if err != nil {
		return decodedImage{}, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return decodedImage{}, errBadPNG
	}
	result := decodedImage{
		info:   info,
		filter: "FlateDecode",
	}
	pixelCount := info.width * info.height
	switch info.components {
	case imageComponentsGray:
		result.data = make([]byte, 0, pixelCount)
	case imageComponentsGrayAlpha:
		result.data = make([]byte, 0, pixelCount)
		result.alphaData = make([]byte, 0, pixelCount)
	case imageComponentsRGB:
		result.data = make([]byte, 0, pixelCount*3)
	case imageComponentsRGBA:
		result.data = make([]byte, 0, pixelCount*3)
		result.alphaData = make([]byte, 0, pixelCount)
	}
	allOpaque := true
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			switch info.components {
			case imageComponentsGray:
				gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
				result.data = append(result.data, gray.Y)
			case imageComponentsGrayAlpha:
				grayAlpha := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				result.data = append(result.data, grayFromNRGBA(grayAlpha))
				result.alphaData = append(result.alphaData, grayAlpha.A)
				allOpaque = allOpaque && grayAlpha.A == 0xFF
			case imageComponentsRGB:
				nrgba := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				result.data = append(result.data, nrgba.R, nrgba.G, nrgba.B)
			case imageComponentsRGBA:
				nrgba := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				result.data = append(result.data, nrgba.R, nrgba.G, nrgba.B)
				result.alphaData = append(result.alphaData, nrgba.A)
				allOpaque = allOpaque && nrgba.A == 0xFF
			}
		}
	}
	if allOpaque {
		result.alphaData = nil
		switch result.info.components {
		case imageComponentsGrayAlpha:
			result.info.components = imageComponentsGray
		case imageComponentsRGBA:
			result.info.components = imageComponentsRGB
		}
	}
	return result, nil
}

func grayFromNRGBA(c color.NRGBA) byte {
	r := float64(c.R)
	g := float64(c.G)
	b := float64(c.B)
	return byte((0.299 * r) + (0.587 * g) + (0.114 * b) + 0.5)
}

func decodeImage(data []byte) (decodedImage, error) {
	if isJPEG(data) {
		info, err := jpegInfo(data)
		if err != nil {
			return decodedImage{}, err
		}
		return decodedImage{
			info:   info,
			data:   data,
			filter: "DCTDecode",
		}, nil
	}
	if isPNG(data) {
		return decodePNG(data)
	}
	return decodedImage{}, errUnsupportedImageFormat
}
