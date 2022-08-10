package demo

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloadZipFile(name string) string {
	config := LoadConfig()
	license := config.License
	url := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz", license)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	filename := fmt.Sprintf("%s.tar.gz", name)
	out, _ := os.Create(filename)
	defer out.Close()
	io.Copy(out, resp.Body)
	return filename
}

func unzipSource(gzipStream io.Reader) string {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Println("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)
	folderName := ""
	check := false
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if !check {
			check = true
			folderName = header.Name
		}
		if err != nil {
			log.Printf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				log.Printf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				log.Printf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Printf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Printf(
				"ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return folderName
}
