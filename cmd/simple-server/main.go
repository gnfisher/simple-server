package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			return
		}

		io.WriteString(w, string(body))
	})

	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f, err := os.Open("file.txt")
		if err != nil {
			log.Fatal(err)
			return
		}
		defer f.Close()
		tr := io.TeeReader(f, w)
		var err2 error
		for err2 == nil {
			_, err2 := tr.Read(make([]byte, 4096))
			if err2 != nil {
				log.Fatal(err)
				return
			}
			time.Sleep(time.Second)

			w.(http.Flusher).Flush()
		}
	})

	http.HandleFunc("/yolo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f, err := os.Open("file.txt")
		if err != nil {
			log.Fatal(err)
			return
		}
		defer f.Close()
		tr := io.TeeReader(f, w)
		sc := bufio.NewScanner(tr)
		sc.Scan()
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}

func NewReader(r io.Reader) io.Reader {
	return &sseReader{sc: bufio.NewScanner(r)}
}

type sseReader struct {
	sc  *bufio.Scanner
	buf []byte
	err error
}

func (r *sseReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	r.refill()
	n = copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, r.err
}

// refill refills the buffer with the next line of data,
// if necessary.
func (r *sseReader) refill() {
	for len(r.buf) == 0 {
		if !r.sc.Scan() {
			if r.err = r.sc.Err(); r.err == nil {
				r.err = io.EOF
			}
			return
		}
		buf := r.sc.Bytes()
		time.Sleep(2 * time.Second)
		r.buf = buf[6:] // remove data: prefix
	}
}
