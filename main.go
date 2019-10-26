package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Client
var dbError error
var mutex = &sync.Mutex{}

//router := mux.NewRouter()
var router *mux.Router

func init() {
	db, dbError = setUpDb()
	if dbError != nil {
		fmt.Printf(dbError.Error())
	}
	router = mux.NewRouter()
}

func main() {
	//PUBLIC
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/login", login).Methods("GET")
	router.HandleFunc("/register", register).Methods("GET")
	router.HandleFunc("/register", signUp).Methods("POST")
	router.HandleFunc("/checkLogin", checkLogin).Methods("POST")
	router.HandleFunc("/user/{userId}", getUser).Methods("GET")
	//INTERNAL
	router.HandleFunc("/user-internal/{email}", getUserByEmail).Methods("PUT")
	router.HandleFunc("/user-data-internal/{userId}", getUserDataByID).Methods("GET")
	router.HandleFunc("/user-internal/{userId}/add-collection", addCollection).Methods("POST")
	router.HandleFunc("/user-internal/{userId}/add-images", uploadImages).Methods("POST")
	router.HandleFunc("/user-internal/{userId}/get-collections", getImageCollections).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.PathPrefix("/node_modules/").Handler(http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules"))))
	fmt.Println("**************STARTING THE SERVER**************")
	err := http.ListenAndServe(":8080", router)
	fmt.Println(err)
}
func index(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/templates/index.html"))
	t.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/templates/login.html")
	t.Execute(w, nil)
}
func register(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/templates/register.html")
	t.Execute(w, nil)
}
func getUser(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/templates/user-page.html")
	err := t.Execute(w, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
func getUserByEmail(w http.ResponseWriter, r *http.Request) {
	var user User
	users := getCollection("users")
	vars := mux.Vars(r)
	email := vars["email"]
	users.FindOne(context.TODO(), bson.D{{"Email", email}}).Decode(&user)
	v, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "failed to marshal", 400)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(v)
}
func getUserDataByID(w http.ResponseWriter, r *http.Request) {
	userID := (mux.Vars(r))["userId"]
	var userData []UserData
	userDataCollection := getCollection("userData")
	filter := bson.D{{"id", userID}}
	cur, err := userDataCollection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("error in finding user data: ", err.Error())
		http.Error(w, "error", 400)
	}
	cur.All(context.TODO(), &userData)
}
func uploadImages(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---UPLOADING IMAGES---")
	err := r.ParseMultipartForm(1200000)
	numberOfFiles, err := strconv.Atoi((r.Form)["fileCount"][0])
	collectionID := (r.Form)["collectionID"][0]

	images := getCollection("images")
	if err != nil {
		fmt.Println("FAILED TO PARSE FORM")
		http.Error(w, err.Error(), 400)
	}
	file := make([]multipart.File, numberOfFiles)
	fileData := make([][]byte, numberOfFiles)
	headers := make([]*multipart.FileHeader, numberOfFiles)
	var operations []mongo.WriteModel
	var wg sync.WaitGroup
	for i := 0; i < numberOfFiles; i++ {
		file[i], headers[i], err = r.FormFile("file-" + strconv.Itoa(i))
		wg.Add(1)
		go performDbWrite(file[i], fileData[i], collectionID, headers[i], &operations, &wg)
	}
	wg.Wait()
	_, err = images.BulkWrite(context.TODO(), operations)
	if err != nil {
		fmt.Println("error: " + err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, "", 200)
}
func performDbWrite(file multipart.File, fileData []byte, collectionID string, header *multipart.FileHeader, operations *[]mongo.WriteModel, wg *sync.WaitGroup) {
	defer file.Close()
	var compressedFile bytes.Buffer
	zw := gzip.NewWriter(&compressedFile)
	operation := mongo.NewInsertOneModel()
	fileSize := header.Size
	fileData = make([]byte, fileSize)
	_, err := file.Read(fileData)
	if err != nil {
		fmt.Println("failed to save file")
		return
	}
	_, err = zw.Write(fileData)
	if err != nil {
		fmt.Println("failed to save file")
		return
	}
	if err := zw.Close(); err != nil {
		fmt.Println("failed to save file")
		return
	}
	image := ImageObj{CollectionID: collectionID, Image: compressedFile.Bytes()}
	operation.SetDocument(image)
	mutex.Lock()
	*operations = append(*operations, operation)
	mutex.Unlock()
	wg.Done()
}

func addCollection(w http.ResponseWriter, r *http.Request) {
	//TODO:verify that the user exists
	var collections []Collection
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&collections)
	userDataCollection := getCollection("userData")
	userID := (mux.Vars(r))["userId"]
	filter := bson.D{{"UserID", userID}}
	update := bson.D{{"$set", bson.D{{"Collections", &collections}}}}
	_, err := userDataCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, "Success", 201)
}

func getImageCollections(w http.ResponseWriter, r *http.Request) {
	userID := (mux.Vars(r))["userId"]
	userData := getCollection("userData")
	userDataObj := UserData{}
	filter := bson.D{{"UserID", userID}}
	userData.FindOne(context.TODO(), filter).Decode(&userDataObj)
	v, err := json.Marshal(userDataObj)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.Write(v)
}
