package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


const testString = "aion nio aia eano nona anio"


func main() {
    app := tview.NewApplication()
    textView := tview.NewTextView().
        SetDynamicColors(true).
        SetRegions(true).
        SetChangedFunc(func() {
            app.Draw()
        })

    colorMap := make([]string, len(testString))
    for i := 0; i < len(testString); i++ {
        colorMap[i] = "white"
    }

    drawText(textView, colorMap)

    currentChar := 0
    textView.Highlight("0")
    app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if currentChar >= len(testString) {
            return event
        }

        if event.Rune() == rune(testString[currentChar]) {
            colorMap[currentChar] = "green"
        } else {
            colorMap[currentChar] = "red"
        }

        drawText(textView, colorMap)
        textView.Highlight(fmt.Sprintf("%d", currentChar+1))
        currentChar += 1

        return event
    })

    textView.SetBorder(true)
    if err := app.SetRoot(textView, true).SetFocus(textView).Run(); err != nil {
        panic(err)
    }
}


func drawText(textView *tview.TextView, colorMap []string) {
    textView.Clear()

    for i, char := range testString {
        fmt.Fprintf(textView, `["%d"][%s]%c[""]`, i, colorMap[i], char)
    }        
} 
