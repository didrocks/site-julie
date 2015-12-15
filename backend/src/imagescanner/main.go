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
	"sync"

	_ "golang.org/x/image/webp"
)

// GlobalPath is used for global path formatting
type GlobalPath struct {
	InputPath        string
	FinaleOutputPath string
	OutputPath       string
	RelativeWebPath  string
}

var globalPaths GlobalPath

/*
 * contains channels
 */
type orchestrer struct {
	wg       *sync.WaitGroup
	receiver <-chan string
}

/*
 * scan input directory. Use a closure for channel communication
 */
func (orch *orchestrer) scanInputDirFunc() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		orch.wg.Add(1)
		if info == nil {
			return fmt.Errorf("%s: does not exists or is not readable ", path)
		}
		translatedPath := strings.Replace(path, globalPaths.InputPath, globalPaths.OutputPath, 1)
		if info.IsDir() {
			err := os.MkdirAll(translatedPath, info.Mode())
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
	globalPaths.FinaleOutputPath = path.Clean(paths[1])
	globalPaths.OutputPath = globalPaths.FinaleOutputPath + ".new"
	webRootPath := path.Clean(paths[2])

	if commonPathIndex := strings.LastIndex(globalPaths.FinaleOutputPath, webRootPath); commonPathIndex == -1 {
		log.Fatal(webRootPath, " needs to be a subdirectory of ", globalPaths.FinaleOutputPath)
	} else {
		globalPaths.RelativeWebPath = globalPaths.FinaleOutputPath[len(webRootPath)+1:]
	}

	if err := os.RemoveAll(globalPaths.OutputPath); err != nil {
		log.Fatal("Couldn't remove ", globalPaths.OutputPath)
	}

	/*
	 *  Init and create our orchestration object and receiver routine
	 */
	orch := &orchestrer{receiver: make(chan string)}

	err := filepath.Walk(globalPaths.InputPath, orch.scanInputDirFunc())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while scanning ", err)
		os.Exit(1)
	}

	// wait for all scanning to be done
	orch.wg.Wait()

	// TODO: tell our receiver routine to write the file

	/*
	 * Save previous generated sets to a backup directory. Try to keep one version.
	 */
	backupDir := globalPaths.FinaleOutputPath + ".bak"
	if err := os.RemoveAll(backupDir); err != nil {
		log.Fatal("Couldn't remove ", globalPaths.OutputPath)
	}
	if err := os.Rename(globalPaths.FinaleOutputPath, backupDir); err != nil {
		log.Fatal("Couldn't archive ", globalPaths.FinaleOutputPath, ". Keeping previous generation around. Newly ",
			"generated content is still available at ", globalPaths.FinaleOutputPath)
	}

	// put in place newly images
	if err := os.Rename(globalPaths.OutputPath, globalPaths.FinaleOutputPath); err != nil {
		log.Fatal("Couldn't save new ", globalPaths.OutputPath, ". Generated content is still available at ",
			globalPaths.FinaleOutputPath, ". Trying to restore old images.")
		if err := os.Rename(backupDir, globalPaths.FinaleOutputPath); err != nil {
			log.Fatal("/!\\ Couldn't restore previous version of ", globalPaths.RelativeWebPath, ". No images are served!")
		}
	}
}
