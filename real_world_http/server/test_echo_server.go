package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/k0kubun/pp"
)

func handler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Print(string(dump))
	pp.Printf("URL: %s\n", r.URL.String())
	pp.Printf("Query: %v\n", r.URL.Query())
	pp.Printf("Method: %s\n", r.Proto)
	pp.Printf("Method: %s\n", r.Method)
	pp.Printf("Header: %v\n", r.Header)
	fmt.Fprintf(w, "<html><body>hello</body></html>\n")
}

func handlerDigest(w http.ResponseWriter, r *http.Request) {
	pp.Printf("URL: %s\n", r.URL.String())
	pp.Printf("Query: %v\n", r.URL.Query())
	pp.Printf("Method: %s\n", r.Proto)
	pp.Printf("Method: %s\n", r.Method)
	pp.Printf("Header: %v\n", r.Header)
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("--body--\n%s\n", string(body))
	if _, ok := r.Header["Authorization"]; !ok {
		w.Header().Add("WWW-Authenticate", `Digest realm="Secret Zone, nonce="TgLc25U2BQA=f510a2780473e186587be702c2e67fe2b04afd", algorithm=MD5, qop="auth"`)
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		fmt.Fprintf(w, "<html><body>secret page</body></html>\n")
	}
}

func handlerUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Connection") != "Upgrade" || r.Header.Get("Upgrade") != "MyProtocol" {
		w.WriteHeader(400)
		return
	}
	fmt.Println("Upgrade to MyProtocol")

	hijacker := w.(http.Hijacker)

	conn, readWriter, err := hijacker.Hijack()
	if err != nil {
		panic(err)
		return
	}

	defer conn.Close()

	response := http.Response{
		StatusCode: 101,
		Header:     make(http.Header),
	}

	response.Header.Set("Upgrade", "MyProtocol")
	response.Header.Set("Connection", "Upgrade")
	response.Write(conn)

	for i := 1; i <= 10; i++ {
		fmt.Fprintf(readWriter, "%d\n", i)
		fmt.Println("->", i)
		readWriter.Flush()
		recv, err := readWriter.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		fmt.Printf("<- %s", string(recv))
		time.Sleep(500 * time.Millisecond)
	}
}

func handlerChunkedResponse(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("expexted http.ResponseWrite to be an http.Flusher")
	}

	for i := 1; i <= 10; i++ {
		fmt.Fprintf(w, "Chunk #%d\n", i)
		flusher.Flush()
		time.Sleep(500 * time.Millisecond)
	}

	flusher.Flush()
}

func main() {
	var httpServer http.Server
	http.HandleFunc("/", handler)
	http.HandleFunc("/digest", handlerDigest)
	http.HandleFunc("/chunked", handlerChunkedResponse)

	log.Println("start http listening :18888")
	httpServer.Addr = ":18888"
	log.Println(httpServer.ListenAndServe())
}
