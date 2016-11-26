package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	schemeHost      = "http://localhost"
	listen          = ":9154"
	uApiGet         = "/api/v1/get/"
	uApiNew         = "/api/v1/new"
	uGet            = "/g"
	uInfo           = "/i"
	uFav            = "/favicon.ico"
	uCss            = "/custom.css"
	maxData         = 1048576 // 1MB
	defaultValidity = 7       // days
	expiryCheck     = 30      // minutes
)

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s \"%s %s %s\" \"%s\"", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, r.Header.Get("User-Agent"))
		handler.ServeHTTP(w, r)
	})
}

func main() {
	store := make(secretStore)
	store.NewEntry("secret", 100, 0, "_authtoken_", "test")
	go store.Expiry()

	auth := makeTokenDB()

	tView := template.New("view")
	tView.Parse(htmlMaster)
	tView.Parse(htmlView)
	tViewErr := template.New("viewErr")
	tViewErr.Parse(htmlMaster)
	tViewErr.Parse(htmlViewErr)
	tViewInfo := template.New("viewInfo")
	tViewInfo.Parse(htmlMaster)
	tViewInfo.Parse(htmlViewInfo)

	http.HandleFunc(uApiGet, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len(uApiGet):]
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if entry, ok := store.GetEntryInfo(id); !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "{}")
		} else {
			store.Click(id)
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(entry); err != nil {
				panic(err)
			}
		}
	})

	http.HandleFunc(uApiNew, func(w http.ResponseWriter, r *http.Request) {
		var entry StoreEntry
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxData))
		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.Unmarshal(body, &entry); err != nil {
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		if !auth.isAuthorized(&entry) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, `{"error":"unauthorized"}`)
		} else {
			id := store.AddEntry(entry, "")
			newEntry, _ := store.GetEntryInfoHidden(id)
			log.Println("New ID:", id)
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(newEntry); err != nil {
				panic(err)
			}
		}
	})

	http.HandleFunc(uGet, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if entry, ok := store.GetEntryInfo(id); !ok {
			w.WriteHeader(http.StatusNotFound)
			tViewErr.ExecuteTemplate(w, "master", nil)
		} else {
			store.Click(id)
			w.WriteHeader(http.StatusOK)
			tView.ExecuteTemplate(w, "master", entry)
		}
	})

	http.HandleFunc(uInfo, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if entry, ok := store.GetEntryInfo(id); !ok {
			w.WriteHeader(http.StatusNotFound)
			tViewErr.ExecuteTemplate(w, "master", nil)
		} else {
			w.WriteHeader(http.StatusOK)
			tViewInfo.ExecuteTemplate(w, "master", entry)
		}
	})

	http.HandleFunc(uFav, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.WriteHeader(http.StatusOK)
		w.Write(favicon)
	})

	http.HandleFunc(uCss, func(w http.ResponseWriter, r *http.Request) {
		css := tryReadFile(cssFileName)
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write(css)
	})

	log.Fatal(http.ListenAndServe(listen, Log(http.DefaultServeMux)))
}
