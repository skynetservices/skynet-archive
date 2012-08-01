package main

import (
  "net/http"
  "html/template"
)

func main(){
  s := &http.Server{
    Addr:           ":8080",
  }


  http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("tmpl"))))
  http.HandleFunc("/", index)

  s.ListenAndServe()
}

func index(w http.ResponseWriter, r *http.Request){
  index := template.Must(template.ParseFiles("tmpl/index.html.template"))

  index.Execute(w, r.URL.Path)
}
