package cli

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func build(args []string) {
	if validAbsoluteDirPathRegex.MatchString(args[0]) {
		// recursively walk through the directory
		filepath.Walk(args[0], func(path string, info os.FileInfo, err error) error {
			// skip directories
			if info.IsDir() {
				return nil
			}

			// if it's a go source file
			if strings.HasSuffix(path, ".go") {
				// build the file
				cmd := exec.Command("go", "build", "-o", getSO(path), "-buildmode=plugin", path)
				err = cmd.Run()
				if err != nil {
					log.Fatal("Could not build ", path)
					return err
				}

				log.Println("Built file: " + getSO(path))
			}

			return nil
		})
	} else {
		log.Fatal("The specified path does not lead to a directory.")
	}
}

// gets the so name of the file
func getSO(path string) (sopath string) {
	return path[:len(path)-2] + "so"
}
