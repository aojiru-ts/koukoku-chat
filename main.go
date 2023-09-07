package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"sync"
	"regexp"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func removeANSI(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(input, "")
}

var allText []string

func main() {
	config := tls.Config{Certificates: []tls.Certificate{}, InsecureSkipVerify: false}
	conn, err := tls.Dial("tcp", "koukoku.shadan.open.ad.jp:992", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	// chat mode
	fmt.Fprintln(conn, "nobody")

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
		var currentMessage []string
		accumulating := false // multi line
	
		for scanner.Scan() {
			line := removeANSI(scanner.Text())
			if strings.HasPrefix(line, ">>") {
				// start
				accumulating = true
				currentMessage = append(currentMessage, line)
			} else if strings.HasSuffix(line, "<<") {
				// end
				currentMessage = append(currentMessage, line)
				joinedMessage := strings.Join(currentMessage, "\n") + "\n"
				allText = append([]string{joinedMessage}, allText...)
				textView.SetText(strings.Join(allText, "\n"))
				// reset
				accumulating = false
				currentMessage = nil
			} else if accumulating {
				// continue
				currentMessage = append(currentMessage, line)
			}
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
