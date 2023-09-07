package util

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloadExportFile(urlFile string, name string) {
	url := fmt.Sprintf(urlFile+"/%s.json.gz", name)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	filename := fmt.Sprintf("%s.json.gz", name)
	out, _ := os.Create(filename)
	defer out.Close()
	io.Copy(out, resp.Body)
}

func Unzip(name string) {
	// Open compressed file
	gzipFile, err := os.Open(name + ".json.gz")
	if err != nil {
		log.Fatal(err)
	}

	// Create a gzip reader on top of the file reader
	// Again, it could be any type reader though
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		log.Fatal(err)
	}
	defer gzipReader.Close()

	// Uncompress to a writer. We'll use a file writer
	outfileWriter, err := os.Create(name + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer outfileWriter.Close()

	// Copy contents of gzipped file to output file
	_, err = io.Copy(outfileWriter, gzipReader)
	if err != nil {
		log.Fatal(err)
	}

	RemoveFile(name + ".json.gz")
}

func RemoveFile(name string) {
	e := os.Remove(name)
	if e != nil {
		log.Fatal(e)
	}
}
