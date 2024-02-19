package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/chrissxMedia/cm3.go"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	outputDir := "/var/bucket"
	if len(os.Args) > 1 {
        // TODO: check whether it exists
		outputDir = os.Args[1]
	}
	infoFile := outputDir + "/_info.csv"

	fmt.Println("STARTING BUcKET, output dir:", outputDir)

	var gfl sync.Mutex

	var requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bucket_requests",
		Help: "Requests",
	}, []string{"remote", "user_agent"})
	var responses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bucket_responses",
		Help: "Responses",
	}, []string{"code"})
	var invalid = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bucket_invalid",
		Help: "Non-POST requests",
	}, []string{"remote", "user_agent"})
	cm3.HandleMetrics(requests, responses, invalid)

	cm3.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			invalid.WithLabelValues(r.RemoteAddr, r.UserAgent()).Inc()
			w.WriteHeader(400)
			fmt.Fprint(w, "Use POST!")
			return
		}
		requests.WithLabelValues(r.RemoteAddr, r.UserAgent()).Inc()
		id := uuid.NewString()
		binFile := outputDir + "/" + id
		gfl.Lock()
		defer gfl.Unlock()
		_, err := os.Open(binFile)
        // TODO: check that it exists, not that we just can't open it
		if err == nil {
			//TODO: regen
			responses.WithLabelValues("500").Inc()
			w.WriteHeader(500)
			fmt.Fprint(w, "BLOB ", id, " already exists.")
			return
		}
		bin, err := os.OpenFile(binFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			responses.WithLabelValues("500").Inc()
			w.WriteHeader(500)
			fmt.Fprint(w, "Can't create bin file: ", err)
			return
		}
		defer bin.Close()
		info, err := os.OpenFile(infoFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			responses.WithLabelValues("500").Inc()
			w.WriteHeader(500)
			fmt.Fprint(w, "Can't open info file: ", err)
			return
		}
		defer info.Close()
		_, err = io.Copy(bin, r.Body)
		if err != nil {
			responses.WithLabelValues("500").Inc()
			w.WriteHeader(500)
			fmt.Fprint(w, "Can't write bin file: ", err)
			return
		}
        // TODO: consider adding r.Header.Get("X-Real-IP")
		// (TODO: check whether r.RemoteAddr is the real IP or the proxy's IP)
		// (docs don't really make it clear)
		csvLine := []string{id, r.RemoteAddr, r.UserAgent(), time.Now().Format(time.RFC3339)}
		csv := csv.NewWriter(info)
		err = csv.Write(csvLine)
		if err != nil {
            // TODO: rethink this error handling
			log.Println("Can't write CSV:", err, "(", csvLine, ")")
		}
		csv.Flush()
		err = csv.Error()
		if err != nil {
			log.Println("Can't write CSV:", err, "(", csvLine, ")")
		}
		responses.WithLabelValues("200").Inc()
		w.WriteHeader(200)
		log.Println("Success:", csvLine)
	})

	cm3.ListenAndServeHttp(":8022", nil)
}
