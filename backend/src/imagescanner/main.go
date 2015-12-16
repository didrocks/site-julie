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

// globalPath is used for global path formatting
type globalPath struct {
	inputPath        string
	finaleOutputPath string
	outputPath       string
	relativeWebPath  string
}

var globalPaths globalPath

/*
 * contains channels
 */
type orchestrer struct {
	wg            *sync.WaitGroup
	collectorChan chan<- string
	collectorDone chan bool
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

	globalPaths.inputPath = path.Clean(paths[0])
	globalPaths.finaleOutputPath = path.Clean(paths[1])
	globalPaths.outputPath = globalPaths.finaleOutputPath + ".new"
	webRootPath := path.Clean(paths[2])

	if commonPathIndex := strings.LastIndex(globalPaths.finaleOutputPath, webRootPath); commonPathIndex == -1 {
		log.Fatal(webRootPath, " needs to be a subdirectory of ", globalPaths.finaleOutputPath)
	} else {
		globalPaths.relativeWebPath = globalPaths.finaleOutputPath[len(webRootPath)+1:]
	}

	if err := os.RemoveAll(globalPaths.outputPath); err != nil {
		log.Fatal("Couldn't remove ", globalPaths.outputPath)
	}

	/*
	 *  Init and create our orchestration object and receiver routine
	 */
	orch := &orchestrer{wg: &sync.WaitGroup{}, collectorDone: make(chan bool)}
	orch.collectorChan = generateCollector(orch.collectorDone)
	orch.collectorChan <- "string"

	// walk through the filesystem, sending channels to it
	if err := os.Chdir(globalPaths.inputPath); err != nil {
		log.Fatal("Couldn't chdir to ", globalPaths.inputPath)
	}
	if err := filepath.Walk(".", orch.scanInputDirFunc()); err != nil {
		fmt.Fprintln(os.Stderr, "Error while scanning ", err)
		os.Exit(1)
	}

	// wait for all scanning to be done
	orch.wg.Wait()

	// tell our receiver that we are done scanning, so that it writes the file and waits for the completion
	close(orch.collectorChan)
	<-orch.collectorDone

	/*
	 * Save previous generated sets to a backup directory. Try to keep one version.
	 */
	backupDir := globalPaths.finaleOutputPath + ".bak"
	if err := os.RemoveAll(backupDir); err != nil {
		log.Fatal("Couldn't remove ", globalPaths.outputPath)
	}
	if _, err := os.Stat(globalPaths.finaleOutputPath); err == nil {
		if err := os.Rename(globalPaths.finaleOutputPath, backupDir); err != nil {
			log.Fatal("Couldn't archive ", globalPaths.finaleOutputPath, ". Keeping previous generation around. Newly ",
				"generated content is still available at ", globalPaths.finaleOutputPath)
		}
	}

	// put in place newly images
	if err := os.Rename(globalPaths.outputPath, globalPaths.finaleOutputPath); err != nil {
		log.Fatal("Couldn't save new ", globalPaths.outputPath, ". Generated content is still available at ",
			globalPaths.finaleOutputPath, ". Trying to restore old images.")
		if err := os.Rename(backupDir, globalPaths.finaleOutputPath); err != nil {
			log.Fatal("/!\\ Couldn't restore previous version of ", globalPaths.relativeWebPath, ". No images are served!")
		}
	}
}
