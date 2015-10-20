/*
 * This command generate json metadata and scale images at different breakpoints for the website. It can be cronned.
 * $1 is the source path, $2 the output one (changes will happen in $2.new before renaming to $2)
 */

package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"imagescanner"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

type globalpath struct {
	inputPath        string
	outputPath       string
	relativeRootPath string
}

var globalPath globalpath

/*
 * Handle input files
 */
func visitInputDir(path string, f os.FileInfo, err error) error {
	if f == nil {
		return fmt.Errorf("%s: does not exists or is not readable ", path)
	}
	translatedPath := strings.Replace(path, globalPath.inputPath, globalPath.outputPath, 1)
	if f.IsDir() {
		err := os.MkdirAll(translatedPath, f.Mode())
		if err != nil {
			log.Fatal("Couldn't create ", translatedPath)
		}
		return nil
	}

	cropheight := 0
	if strings.Contains(translatedPath, "banners") {
		cropheight = imagescanner.BANNERHEIGHT
	}
	img := imagescanner.ImageInputInfo{InURL: path, OutURL: translatedPath, Cropheight: cropheight}
	img.ProcessImage(globalPath.relativeRootPath, globalPath.outputPath)

	return nil
}

/*
 * print command line usage
 */
func usage() {
	fmt.Fprintln(os.Stderr, "Usage: ", os.Args[0], " [OPTIONS] inputdir outputdir relativerootpath")
	flag.PrintDefaults()
}

/*
 * Handle input and parameters
 */
func main() {
	flag.Usage = usage
	flag.Parse()
	paths := flag.Args()

	if len(paths) != 3 {
		flag.Usage()
		os.Exit(1)
	}

	globalPath.inputPath = path.Clean(paths[0])
	globalPath.outputPath = path.Clean(paths[1])
	commonPath := filepath.Base(path.Clean(paths[2]))

	index := strings.LastIndex(globalPath.outputPath, commonPath)
	if index == -1 {
		log.Fatal(globalPath.outputPath, "and", path.Clean(paths[2]), " doesn't have any common directory. Exiting")
	}
	globalPath.relativeRootPath = filepath.Join(path.Clean(paths[2]), globalPath.outputPath[index+len(commonPath):])

	if _, err := os.Stat(globalPath.outputPath); err == nil {
		log.Fatal(globalPath.outputPath, " already exists. Exiting")
	}

	err := filepath.Walk(globalPath.inputPath, visitInputDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while visiting ", err)
		os.Exit(1)
	}
}
