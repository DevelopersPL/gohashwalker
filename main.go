package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
)

var (
	buildtime string
	buildver  string
	files     []File
)

func main() {
	app := cli.NewApp()
	app.Name = "gohashwalker"
	app.Usage = "Point the walker to the starting directory"
	app.Version = buildver + " built " + buildtime
	app.Author = "Daniel Speichert"
	app.Email = "daniel@speichert.pl"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "trim, t", Value: "", Usage: "Trim output filename"},
	}

	app.Action = func(c *cli.Context) {
		go func() {
			if err := filepath.Walk(os.Args[1], walkFn); err != nil {
				log.Fatalf(err.Error())
			}

			// print out JSON file list
			json, err := json.MarshalIndent(files, "", `   `)
			if err == nil {
				fmt.Printf("%s\n", json)
			} else {
				log.Fatalf(err.Error())
			}
			os.Exit(0)
		}()

		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		for {
			select {
			case s := <-sig:
				log.Fatalf("Signal (%d) received, exiting.\n", s)
			}
		}
	}

	app.Run(os.Args)
}

func walkFn(path string, fi os.FileInfo, err error) (e error) {
	if !fi.IsDir() {
		hash, err := hash_file_crc32(path, 0xedb88320)
		if err == nil {
			//fmt.Println(fi.Name(), hash)
			files = append(files, File{
				Path:    path,
				Crc32:   hash,
				Size:    fi.Size(),
				ModTime: fi.ModTime(),
			})
		} else {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func hash_file_crc32(filePath string, polynomial uint32) (crc32 string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	tablePolynomial := crc32.MakeTable(polynomial)
	hash := crc32.New(tablePolynomial)
	if _, err = io.Copy(hash, file); err != nil {
		return
	}
	crc32 = hex.EncodeToString(hash.Sum(nil))
	return
}

type File struct {
	Path    string    `json:"path"`
	Crc32   string    `json:"crc32"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}
