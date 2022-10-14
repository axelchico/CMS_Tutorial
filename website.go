package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/argon2"
)

type Input struct {
	Username string
	password string
}

func (i Input) GetUsername() string {
	return i.Username
}

func (i Input) GetPassword() string {
	return i.password
}

func NewInput(user, passwd string) *Input {
	return &Input{
		Username: user,
		password: passwd,
	}
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	fmt.Println(r.Form) // print information on server side.
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!") // write data to response
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("test.html")
		t.Execute(w, nil)
		http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	} else {
		r.ParseForm()
		// logic part of log in
		tmpUser := fmt.Sprint(r.Form["user"])
		tmpPass := fmt.Sprint(r.Form["pwd"])
		fmt.Printf("Username: %s", tmpUser)
		fmt.Printf("Hashed Password: %s", hashPass(tmpPass))
	}
}

func hashPass(userPass string) string {
	salt := RandStringRunes(16)
	fmt.Println(salt)
	hashedPasswd := argon2.Key([]byte(userPass), []byte(salt), 3, 64*1024, 4, 32)
	encoded := base64.StdEncoding.EncodeToString(hashedPasswd)
	fmt.Printf("%s\n", encoded)
	return encoded
}

func databaseStart() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://axelchico:XdFDThjmSLVwBPEo@cluster0.pmuf2ii.mongodb.net/?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)
	userInfoDatabase := client.Database("userInfo")
	passwordsCollection := userInfoDatabase.Collection("passwords")
	passwordsResult, err := passwordsCollection.InsertOne(ctx, bson.D{
		{Key: "Pass", Value: "test"},
		{Key: "Pass", Value: "Nic Raboy"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(passwordsResult)
	finder, err := passwordsCollection.Find(ctx, bson.M{})
	var allPass []bson.M
	if err = finder.All(ctx, &allPass); err != nil {
		log.Fatal(err)
	}
	for _, passwd := range allPass {
		fmt.Println(passwd["Pass"])
	}
}
func databaseAdd(hashedPass string) {

}

// /Lauches the actual website at port 8080
func StartSite() {
	databaseStart()

	http.HandleFunc("/", sayhelloName)
	http.HandleFunc("/login", login)
	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
