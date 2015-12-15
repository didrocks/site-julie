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
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

// GlobalPath is used for global path formatting
type GlobalPath struct {
	InputPath       string
	OutputPath      string
	RelativeWebPath string
}

var globalPaths GlobalPath

/*
 * Handle input files
 */
func visitInputDir(path string, f os.FileInfo, err error) error {
	if f == nil {
		return fmt.Errorf("%s: does not exists or is not readable ", path)
	}
	translatedPath := strings.Replace(path, globalPaths.InputPath, globalPaths.OutputPath, 1)
	if f.IsDir() {
		err := os.MkdirAll(translatedPath, f.Mode())
		if err != nil {
			log.Fatal("Couldn't create ", translatedPath)
		}
		return nil
	}

	cropheight := 0
	if strings.Contains(translatedPath, "banners") {
		cropheight = BANNERHEIGHT
	}
	img := ImageInputInfo{InURL: path, OutURL: translatedPath, Cropheight: cropheight}
	img.ProcessImage(globalPaths.RelativeWebPath, globalPaths.OutputPath)

	return nil
}

/*
 * print command line usage
 */
func usage() {
	fmt.Fprintln(os.Stderr, "Usage: ", os.Args[0], " [OPTIONS] sourceimagedir outputimagedir websiterootdir")
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

	globalPaths.InputPath = path.Clean(paths[0])
	globalPaths.OutputPath = path.Clean(paths[1])
	webRootPath := path.Clean(paths[2])

	if commonPathIndex := strings.LastIndex(globalPaths.OutputPath, webRootPath); commonPathIndex == -1 {
		log.Fatal(webRootPath, " isn't a sub directory of ", globalPaths.OutputPath, ". Exiting")
	} else {
		globalPaths.RelativeWebPath = globalPaths.OutputPath[len(webRootPath)+1:]
	}

	if _, err := os.Stat(globalPaths.OutputPath); err == nil {
		log.Fatal(globalPaths.OutputPath, " already exists. Exiting")
	}

	err := filepath.Walk(globalPaths.InputPath, visitInputDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while visiting ", err)
		os.Exit(1)
	}
}
