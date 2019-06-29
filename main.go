package main

//ToDo: renew Token.
// gofunction/thread that occasionally retrieves the /secret/tToken
// checks to see if expires soon and renews.

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/gorilla/handlers"
  "github.com/gorilla/mux"
  "io/ioutil"
  "log"
  "net/http"
  "os"
)

type token_auth struct {
  Email string    `json:"email"`
  Password string `json:"password"`
}

func main() {

  rtr := mux.NewRouter()
  rtr.HandleFunc("/tToken", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
      w.WriteHeader(http.StatusUnauthorized)
      fmt.Fprint(w,"{\"response\":\"failed to read body\"}")
      return
    }
    t := token_auth{}
    err = json.Unmarshal(body, &t)
    if err != nil {
      w.WriteHeader(http.StatusUnauthorized)
      fmt.Fprint(w,"{\"response\":\"failed parse json\"}")
      return
    }
    if t.Email == "" {
      w.WriteHeader(http.StatusUnauthorized)
      fmt.Fprint(w,"{\"response\":\"missing email\"}")
      return
    }
    if t.Password == "" {
      w.WriteHeader(http.StatusUnauthorized)
      fmt.Fprint(w,"{\"response\":\"missing password\"}")
      return
    }
    message := map[string]interface{}{
      "grant_type":"password",
      "client_id":"81527cff06843c8634fdc09e8ac0abefb46ac849f38fe1e431c2ef2106796384",
      "client_secret":"c7257eb71a564034f9419ee651c7d0e5f7aa6bfbd18bafb5c5c033b093bb2fa3",
      "email": t.Email,
      "password": t.Password,
    }
    messageBytes, _ :=  json.Marshal(message)
    token, _ := http.Post("https://owner-api.teslamotors.com/oauth/token", "application/json", bytes.NewBuffer(messageBytes))
    w.WriteHeader(token.StatusCode)
    tokenJson, _ := ioutil.ReadAll(token.Body)
    fmt.Fprintf(w,"%s",tokenJson)
  } ).Methods("POST")
  rtr.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
  loggedRouter := handlers.LoggingHandler(os.Stdout, rtr)

  log.Fatal(http.ListenAndServe(":9001", loggedRouter))
}