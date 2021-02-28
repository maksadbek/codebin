package main

import (
	"fmt"
	"os/exec"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"os"
	"context"
	"time"
	"bytes"
)

func compile(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Language string
		Code string
	}{}

	defer r.Body.Close()
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		fmt.Println("here", err)
		return
	}

	tmpfile, err := ioutil.TempFile("", "prog*.go")
	if err != nil {
		fmt.Println(err)
	}

	println(tmpfile.Name())
	defer os.Remove(tmpfile.Name())

	tmpfile.WriteString(payload.Code)
	tmpfile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmds := []string{}

	if payload.Language == "golang" {
		cmds = append(cmds, "go", "run")
	} else if payload.Language == "python" {
		cmds = append(cmds, "python")
	} else {
		fmt.Fprintf(w, "invalid language: %s", payload.Language)
		return
	}

	cmds = append(cmds, tmpfile.Name())
	fmt.Println(">>>", cmds)
	runcmd := exec.CommandContext(ctx, cmds[0], cmds[1:]...)

	encoder := json.NewEncoder(w)

	out := bytes.Buffer{}
	runcmd.Stderr = &out
	runcmd.Stdout = &out
	
	if err := runcmd.Run(); err != nil {
		encoder.Encode(map[string]string{
			"result": out.String(),
		})
		fmt.Println("fuck happened", err, ctx.Err())
		return
	}

	encoder.Encode(map[string]string{
		"result": out.String(),
	})
}

func main() {
	fs := http.FileServer(http.Dir("."))
	
	http.Handle("/", fs)
	http.HandleFunc("/compile", compile)
	http.ListenAndServe(":8080", nil)
}
