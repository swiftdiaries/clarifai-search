package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/bradfitz/slice"
	fetch "github.com/swiftdiaries/clarifai-search/pkg/fetch"
	store "github.com/swiftdiaries/clarifai-search/pkg/store"
)

// URLqueue is used to store the best 10 images
type URLqueue struct {
	ImageURL   string
	ImageValue float64
}

func main() {

	var apikey string
	flag.StringVar(&apikey, "apikey", "a6deb90589324c46847290ed0863e02f", "API KEY used to authorize calls")
	var searchTag string
	flag.StringVar(&searchTag, "tag", "art", "Enter search string")
	redisServer := flag.String("redis", ":6379", "Specify the redis server (e.g. 127.0.0.1:6379)")
	redisPassword := flag.String("redis-password", "", "Specify the redis server password")

	flag.Parse()

	if redisServer != nil && redisPassword != nil {
		store.Pool = store.NewPool(*redisServer, *redisPassword)
	}

	populateResults(apikey)
	//generateResults(searchTag)

	http.HandleFunc("/", resultDisplay)
	http.HandleFunc("/result", output)
	go open("http://localhost:9090/")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func populateResults(apikey string) {
	imageURLs := fetch.TextFiletoURLs()
	var response string //resp is a temporary string

	for _, imageURL := range imageURLs {
		response = store.Get(imageURL)
		if response == "" {
			result := fetch.GetPrediction(apikey, imageURL)
			err := store.Set(imageURL, result)
			if err != nil {
				fmt.Print(err)
			}
		}
	}
}

func generateResults(searchTag string) []URLqueue {
	var response string //resp is a temporary string
	var key = `"name": "train",`
	key = strings.Replace(key, "train", searchTag, 1)
	imageURLs := fetch.TextFiletoURLs()
	var queue []URLqueue

	for _, URL := range imageURLs {
		response = store.Get(URL)
		if strings.Contains(response, key) {
			//fmt.Printf("At position %d: %v \n", i, URL)
			res := strings.Split(response, key)
			splitKey := res[1][:22]
			value, _ := strconv.ParseFloat(strings.Split(res[1], splitKey)[1][:6], 64)
			//fmt.Println(value)
			queue = append(queue, URLqueue{ImageURL: URL, ImageValue: value})
		}
	}

	slice.Sort(queue[:], func(i, j int) bool {
		return queue[i].ImageValue > queue[j].ImageValue
	})

	if len(queue) > 10 {
		queue = queue[:10]
	} else if len(queue) == 0 {
		queue = append(queue, URLqueue{ImageURL: "https://data.whicdn.com/images/5496723/large.png", ImageValue: 0.10})
		fmt.Println("Not found")
	}
	return queue
}

func output(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		queue := generateResults(r.Form["searchtag"][0])
		t, _ := template.ParseFiles("result.html")
		t.Execute(w, queue)
	} else {
		r.ParseForm()

		fmt.Println("Search Tag:", r.Form["searchtag"])
		queue := generateResults(r.Form["searchtag"][0])
		t, _ := template.ParseFiles("result.html")
		t.Execute(w, queue)
		//http.Redirect(w, r, "http://localhost:9090/result", http.StatusSeeOther)
	}
}

func resultDisplay(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("home.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Print(err)
	}

}

func open(url string) error {

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()

}
