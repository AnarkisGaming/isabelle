package crash

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"

	"get.cutie.cafe/isabelle/config"
	"get.cutie.cafe/isabelle/types"
)

// Parse an Exception. Caveat emptor: if Config.Files.File1 ends in .xml, it is parsed as a SerializableException.
func Parse(zipReader *zip.Reader) (string, string, error) {
	var exception *types.SerializableException
	var specs string

	excFmt := ""

	for _, file := range zipReader.File {
		if file.Name == config.Config.Files.File1 || file.Name == config.Config.Files.File2 {
			reader, err := file.Open()
			if err != nil {
				fmt.Printf("Found a file named %s but could not read it: %v\n", file.Name, err)
				return "", "", err
			}
			defer reader.Close()

			fileBytes, err := ioutil.ReadAll(reader)
			if err != nil {
				fmt.Printf("Found a file named %s but could not read it: %v\n", file.Name, err)
				return "", "", err
			}

			if file.Name == config.Config.Files.File1 {
				if strings.HasSuffix(config.Config.Files.File1, ".xml") {
					if err = xml.Unmarshal(fileBytes, &exception); err != nil {
						fmt.Printf("Found a file named %s but could not read it: %v\n", file.Name, err)
						return "", "", err
					}
				} else {
					excFmt = string(fileBytes)
				}
			} else if file.Name == config.Config.Files.File2 {
				specs = string(fileBytes)
			}
		}
	}

	if exception != nil {
		excFmt = exception.Type + ": " + exception.Message + "\n" + exception.StackTrace
	}

	if len(excFmt) > 1500 {
		excFmt = excFmt[0:1500]

		// make traces slightly prettier
		st := strings.Split(excFmt, "\n")
		excFmt = strings.Join(st[0:len(st)-1], "\n")
	}

	return excFmt, specs, nil
}
