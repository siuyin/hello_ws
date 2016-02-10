package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

const addr = ":4000"

//const addr = "localhost:4000"

func main() {
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("listenAndServe:", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, addr)
}

var rootTemplate = template.Must(template.New("root").Parse(`
<!DOCTYPE html>
<html>
  <head>
  <meta charset="utf-8" />
  <script>
    var input, output, websock;

    function showMessage(m) {
      var p = document.createElement("p");
      p.innerHTML = m;
      output.appendChild(p);
    }

    function onMessage(e) {
      showMessage(e.data);
    }

    function onClose() {
      showMessage("Conection closed.");
    }

    function sendMessage() {
      var m = input.value;
      input.value = "";
      websock.send(m);
      showMessage(m);
    }

    function onKey(e) {
      if (e.keyCode == 13) {
        sendMessage();
      }
    }

    function init() {
      input = document.getElementById("input");
      input.addEventListener("keyup", onKey, false);

      output = document.getElementById("output");

      websock = new WebSocket("ws://go-108182.nitrousapp.com:4000/socket");
      websock.onmessage = onMessage;
      websock.onclose = onClose;
    }

    window.addEventListener("load", init, false);
  </script>
  </head>
  <body>
    <input id="input" type="text">
    <div id="output"></div>
  </body>
</html>
`))

type socket struct {
	io.ReadWriter
	done chan bool
}

func (s socket) Close() error {
	s.done <- true
	return nil
}

func socketHandler(c *websocket.Conn) {
	s := socket{c, make(chan bool)}
	go match(s)
	<-s.done
}

var partner = make(chan io.ReadWriteCloser)

func match(c io.ReadWriteCloser) {
	fmt.Fprint(c, "Waiting for partner...")
	select {
	case partner <- c:
	// we are able to send to partner, thus we are done here.
	case p := <-partner:
		// we are receiving from partner, let's chat
		chat(p, c)
	}
}
func chat(a, b io.ReadWriteCloser) {
	fmt.Fprintln(a, "Found one! Say hi.")
	fmt.Fprintln(b, "Found one! Say hi.")
	errc := make(chan error, 1)
	go cp(a, b, errc)
	go cp(b, a, errc)
	if err := <-errc; err != nil {
		log.Fatal("non-nil copy:", err)
	}
	a.Close()
	b.Close()
}
func cp(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}
