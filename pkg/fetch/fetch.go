package fetch

import (
	"bufio"
	"fmt"
	"log"
	"os"

	cl "github.com/mpmlj/clarifai-client-go"
	client "github.com/swiftdiaries/clarifai-search/pkg/client"
)

// ReadFromFile is used to read the list of image URLs from the input.txt file
func ReadFromFile(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return lines
}

// TextFiletoURLs returns an array of URLs from a textfile
func TextFiletoURLs() []string {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	imageURLsfromFile := ReadFromFile(currentWorkingDirectory + "/images.txt")
	return imageURLsfromFile
}

// GetPrediction is used to get prediction from the API for an input image URL
func GetPrediction(apikey string, inputURL string) string {
	var err error
	var clsess *cl.Session

	clsess = client.CreateSession(apikey)

	data := cl.InitInputs()

	_ = data.AddInput(cl.NewImageFromURL(inputURL), "")

	data.SetModel(cl.PublicModelGeneral)

	resp, err := clsess.Predict(data).Do()
	if err != nil {
		panic(err)
	}
	str := cl.PE(resp)
	return str
}
