package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Client
var dbError error
var mutex = &sync.Mutex{}
var mutexRetrieve = &sync.Mutex{}

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
	router.HandleFunc("/register", signUp).Methods("POST")
	router.HandleFunc("/login", checkLogin).Methods("POST")
	router.HandleFunc("/user/{userId}", getUser).Methods("GET")
	router.HandleFunc("/shared/{userId}/{collectionID}", getSharedCollection).Methods("GET")
	//INTERNAL
	router.HandleFunc("/user-internal/{email}", getUserByEmail).Methods("PUT")
	router.HandleFunc("/user-data-internal/{userId}", getUserDataByID).Methods("GET")
	router.HandleFunc("/user-internal/{userId}/add-collection", addCollection).Methods("POST")
	router.HandleFunc("/user-internal/{userId}/add-images", uploadImages).Methods("POST")
	router.HandleFunc("/user-internal/{userId}/get-collections", getImageCollections).Methods("GET")
	router.HandleFunc("/user-internal/{userId}/get-collections/{collectionId}", getSharedCollectionContent).Methods("GET")
	router.HandleFunc("/user-internal/collections/{collectionId}/images/{imageId}", getImages).Methods("GET")
	router.HandleFunc("/user-internal/remove-collection-images/{collectionId}", removeImages).Methods("DELETE")
	router.HandleFunc("/user-internal/{userID}/remove-image/collections/{collectionID}/images/{imageID}", removeSpecificImages).Methods("DELETE")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.PathPrefix("/node_modules/").Handler(http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules"))))
	fmt.Println("**************STARTING THE SERVER**************")
	err := http.ListenAndServe(GetPort(), router)
	// err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", router)
	fmt.Println(err)
}
func GetPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}
func getSharedCollectionContent(w http.ResponseWriter, r *http.Request) {
	userId := (mux.Vars(r))["userId"]
	collectionID := (mux.Vars(r))["collectionId"]
	var userDataObj UserData
	filter := bson.D{{"UserID", userId}}
	userData := getCollection("userData")
	userData.FindOne(context.TODO(), filter).Decode(&userDataObj)
	for _, collection := range userDataObj.Collections {
		if collection.CollectionID == collectionID {
			v, err := json.Marshal(collection)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			w.Header().Add("Content-type", "application/json")
			w.Write(v)
			return
		}
	}
	http.Error(w, "no collections found!", 404)
}
func getSharedCollection(w http.ResponseWriter, r *http.Request) {
	sharedPage, _ := os.Open("static/templates/shared-collection.html")
	fileSize, err := sharedPage.Stat()
	sharedPageData := make([]byte, fileSize.Size())
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	_, err = sharedPage.Read(sharedPageData)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(sharedPageData)
}

func removeSpecificImages(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here!")
	userID := (mux.Vars(r))["userID"]
	collectionID := (mux.Vars(r))["collectionID"]
	imageID := (mux.Vars(r))["imageID"]
	fmt.Println(collectionID)
	fmt.Println(imageID)
	images := getCollection("images")
	userData := getCollection("userData")
	filter := bson.D{{"ImageID", imageID}}
	_, err := images.DeleteOne(context.TODO(), filter)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	filterUserData := bson.D{{"UserID", userID}}
	// update := bson.D{{"$pull", bson.D{{"Collections.$[t].ImageIDs", bson.D{{"$in", imageID}}}}}}
	update := bson.D{{"$pull", bson.M{"Collections.$[t].ImageIDs": imageID}}}
	_, err = userData.UpdateOne(context.TODO(), filterUserData, update, options.Update().SetArrayFilters(options.ArrayFilters{Filters: []interface{}{bson.M{"t.CollectionID": collectionID}}}))
	// userData.FindOne(context.TODO())
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, "Success", 200)
}
func performImageRemove(imageID, collectionID string, wg *sync.WaitGroup, operations *[]mongo.WriteModel) {
	operation := mongo.NewDeleteOneModel()
	filter := bson.D{{"ImageID", imageID}}
	operation.SetFilter(filter)
	mutex.Lock()
	*operations = append(*operations, operation)
	mutex.Unlock()
	wg.Done()
}
func removeImages(w http.ResponseWriter, r *http.Request) {
	collectionID := (mux.Vars(r))["collectionId"]
	images := getCollection("images")
	filter := bson.D{{"CollectionID", collectionID}}
	_, err := images.DeleteMany(context.TODO(), filter)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, "SUCCESS", 200)
}
func getImages(w http.ResponseWriter, r *http.Request) {
	collectionID := (mux.Vars(r))["collectionId"]
	imageID := (mux.Vars(r))["imageId"]
	images := getCollection("images")
	filter := bson.D{{"CollectionID", collectionID}, {"ImageID", imageID}}
	var imageObj ImageObjRetrieve
	images.FindOne(context.TODO(), filter).Decode(&imageObj)
	f, err := ioutil.TempFile("static/temp", "Image-*."+imageObj.FileType)
	defer func(file *os.File) {
		file.Close()
		os.Remove(file.Name())
	}(f)
	if err != nil {
		fmt.Println("failed: ", err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	f.Write(imageObj.Image)
	w.Header().Set("Content-Disposition", "attachment; filename="+f.Name())
	w.Header().Set("Content-Type", http.DetectContentType(imageObj.Image))
	w.Header().Set("Content-Length", strconv.FormatInt(imageObj.Size, 10))
	f.Seek(0, 0)
	io.Copy(w, f)
}

// func umCompressImageAndAddToForm(imageData ImageObjRetrieve, index int, wg *sync.WaitGroup, imageForm *multipart.Writer) {
// 	f, err := ioutil.TempFile("static/temp", "Image-*."+imageData.FileType)
// 	defer func(file *os.File) {
// 		file.Close()
// 		os.Remove(file.Name())
// 	}(f)
// if err != nil {
// 	fmt.Println("failed: ", err.Error())
// 	return
// }
// 	f.Write(imageData.Image)
// 	mutexRetrieve.Lock()
// 	_, err = imageForm.CreateFormFile("file-"+strconv.Itoa(index), f.Name())
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	mutexRetrieve.Unlock()
// }
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
	userPage, _ := os.Open("static/templates/user-page.html")
	fileSize, err := userPage.Stat()
	userPageData := make([]byte, fileSize.Size())
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	_, err = userPage.Read(userPageData)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(userPageData)
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
		fileID := (r.Form)["file-"+strconv.Itoa(i)+"-id"][0]
		wg.Add(1)
		go performDbWrite(file[i], fileData[i], collectionID, headers[i], &operations, &wg, fileID)
	}
	wg.Wait()
	if len(operations) > 0 {
		_, err = images.BulkWrite(context.TODO(), operations)
	}
	if err != nil {
		fmt.Println("error: " + err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, "", 200)
}
func performDbWrite(file multipart.File, fileData []byte, collectionID string, header *multipart.FileHeader, operations *[]mongo.WriteModel, wg *sync.WaitGroup, fileID string) {
	defer file.Close()
	fileType := header.Filename
	fileType = fileType[strings.Index(fileType, ".")+1:]
	operation := mongo.NewInsertOneModel()
	fileSize := header.Size
	fileData = make([]byte, fileSize)
	_, err := file.Read(fileData)
	if err != nil {
		fmt.Println("failed to save file")
		return
	}
	if err != nil {
		fmt.Println("failed to save file")
		return
	}
	image := ImageObjSend{CollectionID: collectionID, Image: fileData, FileType: fileType, Size: fileSize, ImageID: fileID}
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
