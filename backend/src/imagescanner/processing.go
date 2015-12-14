package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cenkalti/dominantcolor"
	"github.com/disintegration/imaging"
)

// BasePath to produce relative urls
var BasePath string

// ImageInputInfo is some metadata about the source file and constraints to be processed
type ImageInputInfo struct {
	InURL      string
	OutURL     string
	Cropheight int
}

type imageInfo struct {
	/* relative url */
	title       string
	year        int
	otherinfo   string
	relativeURL string
	url         string
	width       int
	height      int
	color       string
}

// ProcessImage reads image by reading info and resize optionally to desiredSize, we just print errors and skip to the next element
func (inputInfo ImageInputInfo) ProcessImage(relativerootpath string, basepath string) {

	/* first, fetch the data */
	f, err := os.Open(inputInfo.InURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while reading image ", inputInfo.InURL, err)
		return
	}
	defer f.Close()

	sourceimg, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't decode image ", inputInfo.InURL, err)
		return
	}

	size := sourceimg.Bounds().Size()
	averagecolor := dominantcolor.Hex(dominantcolor.Find(sourceimg))

	for scalefactor, widths := range IMGBREAKPOINTS {
		for _, width := range widths {
			if width > size.X {
				fmt.Fprintln(os.Stderr, "Couldn't scale back to ", strconv.Itoa(width), " as the image is ", strconv.Itoa(size.X), " width for ", inputInfo.InURL)
				continue
			}
			outputInfo, err := generateImageAndInfos(inputInfo, relativerootpath, sourceimg, width, int(float32(inputInfo.Cropheight)*scalefactor), averagecolor, basepath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "An error occured while extracting or saving ", inputInfo.InURL, err)
				continue
			}
			fmt.Print(outputInfo)
		}
	}

	// Add image (eventually cropped) at full width
	outputInfo, err := generateImageAndInfos(inputInfo, relativerootpath, sourceimg, -1, inputInfo.Cropheight, averagecolor, basepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occured while extracting or saving ", inputInfo.InURL, err)
	}
	fmt.Print(outputInfo)

}

func generateImageAndInfos(inputInfo ImageInputInfo, relativerootpath string, sourceimg image.Image, width int, height int, averagecolor string, basepath string) (imageInfo, error) {
	resultingImg := imageInfo{color: averagecolor}

	// if width is <0, then we want the full image width
	if width < 0 {
		width = sourceimg.Bounds().Max.X
	}

	img := resultingImg.resizeOrCrop(sourceimg, width, height)
	resultingImg.computefilepath(inputInfo.OutURL, relativerootpath, basepath)
	err := resultingImg.save(img)
	if err != nil {
		return resultingImg, err
	}
	return resultingImg, nil
}

func (result *imageInfo) computefilepath(sourceurl string, relativerootpath string, basepath string) {
	filename := path.Base(sourceurl)
	destdir := path.Dir(sourceurl)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	result.url = path.Join(destdir, filename+"_"+strconv.Itoa(result.width)+"_"+strconv.Itoa(result.height)+".jpg")
	relativelocalpath := strings.TrimPrefix(result.url, basepath)
	result.relativeURL = path.Join(relativerootpath, relativelocalpath)
}

/* resize when there is no height constraint, otherwise crop */
func (result *imageInfo) resizeOrCrop(img image.Image, width int, height int) image.Image {
	var imgdata image.Image
	if height == 0 {
		imgdata = imaging.Resize(img, width, height, imaging.Lanczos)
	} else {
		imgdata = imaging.CropCenter(img, width, height)
	}
	result.width = imgdata.Bounds().Max.X
	result.height = imgdata.Bounds().Max.Y
	return imgdata
}

func (result *imageInfo) save(img image.Image) error {
	f, err := os.Create(result.url)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, nil)
}
