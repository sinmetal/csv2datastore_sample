package csv2datastore_sample

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine/datastore"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/admin/importCSV", handlerImportCSV)
	http.HandleFunc("/queue/push2datastore", handlerPushToDatastore)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World"))
}

func handlerImportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bucketName := r.FormValue("bucketName")
	objectName := r.FormValue("objectName")

	err := importCSV(ctx, bucketName, objectName)
	if err != nil {
		log.Errorf(ctx, "Failed to import csv: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func importCSV(ctx context.Context, bucketName string, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create client: %v", err)
	}

	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)

	or, err := object.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create object reader: %v", err)
	}
	defer or.Close()

	r := csv.NewReader(or)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Failed to read file: %v", err)
		}
		log.Infof(ctx, "%v", record)
		line := strings.Join(record, ",")
		taskqueue.Add(ctx, &taskqueue.Task{
			Path:    "/queue/push2datastore",
			Payload: []byte(line),
			Method:  "POST",
		}, "default")
	}

	return nil
}

func handlerPushToDatastore(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	reader := csv.NewReader(r.Body)
	defer r.Body.Close()

	record, err := reader.ReadAll()
	if err != nil {
		log.Errorf(ctx, "Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Infof(ctx, "%v", record)

	id, err := strconv.Atoi(record[0][0])
	if err != nil {
		log.Errorf(ctx, "Failed to convert Atoi id. request body: %v", record[0][0])
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	price, err := strconv.Atoi(record[0][2])
	if err != nil {
		log.Errorf(ctx, "Failed to convert Atoi price. request body: %v", record[0][2])
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := datastore.NewKey(ctx, "Sample", "", int64(id), nil)
	_, err = datastore.Put(ctx, key, &struct {
		Name  string
		Price int
	}{
		Name:  record[0][1],
		Price: price,
	})
	if err != nil {
		log.Errorf(ctx, "Failed to put to Datastore. err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
