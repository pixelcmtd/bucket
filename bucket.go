package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
)

func main() {
	outputDir := "/var/bucket"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}
	infoFile := outputDir + "/_info.csv"

	var gfl sync.Mutex

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(400)
			fmt.Fprint(w, "Use POST!")
			return
		}
		id := uuid.NewString()
		binFile := outputDir + "/" + id
		gfl.Lock()
		_, err := os.Open(binFile)
		if err == nil {
			//TODO: regen
			w.WriteHeader(500)
			fmt.Fprint(w, "BLOB ", id, " already exists.")
			return
		}
		bin, err := os.OpenFile(binFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Can't create bin file: ", err)
			return
		}
		info, err := os.OpenFile(infoFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Can't open info file: ", err)
			bin.Close()
			return
		}
		rdr := io.TeeReader(r.Body, bin)
		for {
			b := make([]byte, 4096)
			if _, err := rdr.Read(b); err != nil {
				break
			}
		}
		remote := r.RemoteAddr
		ua := r.UserAgent()
		csv := csv.NewWriter(info)
		err = csv.Write([]string{id, remote, ua})
		if err != nil {
			fmt.Println("Can't write CSV:", err, "(", id, "from", remote, ":", ua, ")")
		}
		csv.Flush()
		err = csv.Error()
		if err != nil {
			fmt.Println("Can't write CSV:", err, "(", id, "from", remote, ":", ua, ")")
		}
		bin.Close()
		info.Close()
		gfl.Unlock()
		w.WriteHeader(200)
	})

	http.ListenAndServe(":8022", nil)
}
