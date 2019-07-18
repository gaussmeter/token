package main

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
  "time"
)

var (
  client_id = "81527cff06843c8634fdc09e8ac0abefb46ac849f38fe1e431c2ef2106796384"
  client_secret = "c7257eb71a564034f9419ee651c7d0e5f7aa6bfbd18bafb5c5c033b093bb2fa3"
)

type token_auth struct {
  Email string    `json:"email"`
  Password string `json:"password"`
}

type token struct {
  AccessToken string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  TokenType string `json:"token_type"`
  ExpiresIn int64 `json:"expires_in"`
  CreatedAt int64 `json:"created_at"`

}

// Todo: clean this up...
//  it "works for now" but
//  in cases where the token is renewed, but fails to store, it could be lost. (retry post to config)
//  also... capture/check/handle all the errors...
func renewtToken() {
  var token token
  for {
    tokenResponse, _ := http.Get("http://config:8443/secret/tToken")
    responseBody, _ := ioutil.ReadAll(tokenResponse.Body)
    _ = json.Unmarshal(responseBody, &token)
    // 3888000 seconds = 45 Days (value of token.ExpiresIn as time of writing)
    // 86400 seconds = 1 Day
    // Renew token 7 days before ie expires.
    renewAt := token.CreatedAt + token.ExpiresIn - (86400 * 7)
    if time.Now().Unix() > renewAt {
      message := map[string]interface{}{
        "grant_type":    "refresh_token",
        "client_id":     client_id,
        "client_secret": client_secret,
        "refresh_token": token.RefreshToken,
      }
      messageBytes, _ := json.Marshal(message)
      renewTokenResponse, err := http.Post("https://owner-api.teslamotors.com/oauth/token", "application/json", bytes.NewBuffer(messageBytes))
      log.Printf("renew response code: %d", renewTokenResponse.StatusCode)
      if err == nil && renewTokenResponse.StatusCode == http.StatusOK {
        log.Printf("token renewed.")
        renewedToken, _ := ioutil.ReadAll(renewTokenResponse.Body)
        storedToken, err := http.Post("http://config:8443/secret/tToken", "text/html", bytes.NewBuffer(renewedToken))
        if err == nil && storedToken.StatusCode == http.StatusOK {
          log.Printf("token stored")
        } else {
          log.Printf("error: %s", err.Error())
        }
      }
    }
    log.Printf("token expires in %d seconds", renewAt - time.Now().Unix())
    //sleep until tomorrow.
    time.Sleep(86400 * time.Second)
  }
}

func main() {
  go renewtToken()

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
      "client_id": client_id,
      "client_secret": client_secret,
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