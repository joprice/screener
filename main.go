package main

import (
	"bitbucket.org/tebeka/selenium"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strconv"
)

type Screenshot struct {
	Url  string
	Data []byte
}

func main() {
	mongoHost := flag.String("mongo-host", "127.0.0.1", "mongo host")
	httpPort := flag.Int("port", 8080, "http port")
  webDriverUrl := flag.String("web-driver-url", "http://127.0.0.1:4444", "web driver url")

	flag.Parse()

	wd, err := webDriver(*webDriverUrl)
  if err != nil {
    log.Fatal(err)
  }
	defer wd.Quit()

	session, err := mgo.Dial(*mongoHost)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	http.HandleFunc("/screenshot", screenshotHandler(wd, session))

	port := strconv.Itoa(*httpPort)
	fmt.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func webDriver(url string) (selenium.WebDriver, error) {
	caps := selenium.Capabilities{}
	wd, err := selenium.NewRemote(caps, url)
	if err != nil {
    return nil, err
	}
	return wd, err
}

func screenshotHandler(wd selenium.WebDriver, session *mgo.Session) http.HandlerFunc {
	c := session.DB("screener").C("screenshots")

	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			http.Error(w, "parameter 'url' is required", 400)
			return
		}

		var data []byte

		shouldRefresh := r.URL.Query().Get("refresh") == "true"
		if shouldRefresh {
			var err error
			data, err = refresh(wd, c, url)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}

		data, err := getOrRefresh(wd, c, url)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		writeData(w, data)
	}
}

func loadUrl(wd selenium.WebDriver, url string) ([]byte, error) {
	err := wd.Get(url)

	if err != nil {
		return nil, err
	}

	data, err := wd.Screenshot()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func writeData(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func refresh(wd selenium.WebDriver, c *mgo.Collection, url string) ([]byte, error) {
	data, err := loadUrl(wd, url)
	if err != nil {
		return nil, err
	}
	if err := c.Insert(&Screenshot{url, data}); err != nil {
		return nil, err
	}
	return data, nil
}

func getOrRefresh(wd selenium.WebDriver, c *mgo.Collection, url string) ([]byte, error) {
	existing := Screenshot{}
	if err := c.Find(bson.M{"url": url}).One(&existing); err != nil {
		if err == mgo.ErrNotFound {
			return refresh(wd, c, url)
		}
		return nil, err
	}
	return existing.Data, nil
}
