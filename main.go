package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/akamensky/argparse"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type button struct {
	Name  string
	Bytes string // of length 2
}

// Inputs are represented by 3 bytes on /dev/ttyAML3.
// 2 bytes for button ID, then 01 = pressed, 00 = not pressed.
// For sliders the third byte is between 0x00 (lowest) and 0x7f (highest).
// For wheels 0x7f = left , 0x01 = right.
//
// You can find out with strace -e read -e status=successful -fp "$(pidof rc_audio_mixer)" -xx -P /dev/ttyAML3
// (Note: While strace is running audio will be broken, kill it to use the Rødecaster again)
// Do not send commands too quickly! sleep at least 0.01s between pressed and not.
var BUTTONS = []button{
	button{"listen 1", "\xb0\x19"},
	button{"listen 2", "\xb1\x19"},
	button{"listen 3", "\xb2\x19"},
	button{"listen 4", "\xb3\x19"},

	button{"mute 1", "\xb0\x1d"},
	button{"mute 2", "\xb1\x1d"},
	button{"mute 3", "\xb2\x1d"},
	button{"mute 4", "\xb3\x1d"},

	button{"settings 1", "\xb0\x15"},
	button{"settings 2", "\xb1\x15"},
	button{"settings 3", "\xb2\x15"},
	button{"settings 4", "\xb3\x15"},

	button{"page left", "\xb0\x37"},
	button{"page right", "\xb1\x37"},

	button{"SMART Pad 1 (top left)", "\xb0\x23"},
	button{"SMART Pad 2 (middle left)", "\xb1\x23"},
	button{"SMART Pad 3 (bottom left)", "\xb2\x23"},
	button{"SMART Pad 4 (top right)", "\xb3\x23"},
	button{"SMART Pad 5 (middle right)", "\xb4\x23"},
	button{"SMART Pad 6 (bottom right)", "\xb5\x23"},

	button{"Big wheel button", "\xb0\x2f"},
	// big wheel: b0 2b
	// headphone 1 wheel: b1 2b (button non-functional)
	// headphone 1 wheel: b2 2b (button non-functional)

	// Fader 1: b0 0f
	// Fader 2: b1 0f
	// Fader 3: b2 0f
	// Fader 4: b3 0f
}

func calcButtonsByName() map[string]button {
	res := make(map[string]button)
	for _, btn := range BUTTONS {
		res[btn.Name] = btn
	}
	return res
}

var buttonsByName = calcButtonsByName()

//go:embed static/index.html
var INDEX_FILE string

func rootHandler(w http.ResponseWriter, r *http.Request) {
	var buttonNames = make([]string, len(BUTTONS))
	for i, btn := range BUTTONS {
		buttonNames[i] = btn.Name
	}
	buttonJson, err := json.Marshal(buttonNames)
	if err != nil {
		panic(err)
	}
	indexHtml := strings.Replace(INDEX_FILE, "{{ buttonJSON }}", html.EscapeString(string(buttonJson)), -1)

	io.WriteString(w, indexHtml)
}

//go:embed static/client.js
var CLIENT_JS_FILE string

func clientJSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	io.WriteString(w, CLIENT_JS_FILE)
}

func sendByte(fd uintptr, b byte) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCSTI, uintptr(unsafe.Pointer(&b)))
	if errno != 0 {
		msg := fmt.Sprintf("TIOCSTI failed with errno %d", errno)
		return errors.New(msg)
	}
	return nil
}

func pressButton(fd uintptr, btn button) error {
	var err error
	bytes := []byte(btn.Bytes)
	fmt.Printf("bytes in button %s: %x\n", btn.Name, btn.Bytes)

	for _, b := range bytes {
		err = sendByte(fd, b)
		if err != nil {
			return err
		}
	}
	err = sendByte(fd, byte('\x00'))
	if err != nil {
		return err
	}

	// Required: wait some time between button press and release
	time.Sleep(50 * time.Millisecond)

	for _, b := range bytes {
		err = sendByte(fd, b)
		if err != nil {
			return err
		}
	}
	err = sendByte(fd, byte('\x01'))
	if err != nil {
		return err
	}
	return nil
}

func pressButtonsHandler(w http.ResponseWriter, r *http.Request) {
	type requestBodyType struct {
		Buttons []string `json:"buttons"`
	}
	buf := requestBodyType{}

	err := json.NewDecoder(r.Body).Decode(&buf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestButtonNames := buf.Buttons
	buttons := make([]button, 0)
	for _, btnName := range requestButtonNames {
		btn, ok := buttonsByName[btnName]
		if !ok {
			msg := fmt.Sprintf("Cannot find button %s", btnName)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		buttons = append(buttons, btn)
	}

	file, err := os.OpenFile("/dev/ttyAML3", os.O_WRONLY, 0)
	if err != nil {
		log.Printf("Failed to set buttons: cannot terminal: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fd := file.Fd()

	for _, b := range buttons {
		err := pressButton(fd, b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "{}")
}

func main() {
	// Set this program to lowest priority – any audio handling is more important!
	syscall.Setpriority(syscall.PRIO_PGRP, 0, 19)

	parser := argparse.NewParser("rc2http", "HTTP server for Rødecaster Duo")
	installService := parser.Flag("", "install-service", &argparse.Options{Help: "Install and start as a service"})
	port := parser.String("p", "port", &argparse.Options{Default: ":80", Help: "Address & port to listen to"})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	if *installService {
		InstallService()
		os.Exit(0)
	}

	http.HandleFunc("/press-buttons", pressButtonsHandler)
	http.HandleFunc("/static/client.js", clientJSHandler)
	http.HandleFunc("/", rootHandler)

	log.Fatal(http.ListenAndServe(*port, nil))
}
