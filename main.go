package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var allText []string

func main() {
	config := tls.Config{Certificates: []tls.Certificate{}, InsecureSkipVerify: false}
	conn, err := tls.Dial("tcp", "koukoku.shadan.open.ad.jp:992", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	inputField := tview.NewInputField().
		SetLabel("Send: ").
		SetFieldWidth(0)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := inputField.GetText()
			fmt.Fprintln(conn, text)
			inputField.SetText("")
		}
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(inputField, 1, 1, true)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, ">>") {
				continue
			}
			allText = append([]string{line}, allText...)
			textView.SetText(strings.Join(allText, "\n"))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
			log.Fatalf("Error running the app: %s", err)
		}
	}()

	wg.Wait()
}
