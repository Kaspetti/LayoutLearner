package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Kaspetti/LayoutLearner/internal/dictionary"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


func main() {
    rand.Seed(time.Now().UTC().UnixNano())

    characterPriority, err := dictionary.GetCharacterPriority("resources/words.txt")
    if err != nil {
        panic(err)
    }

    words, colorMap := newGame(characterPriority)

    app := tview.NewApplication()
    textView := tview.NewTextView().
        SetDynamicColors(true).
        SetRegions(true)

    drawText(textView, words, colorMap)

    currentChar := 0
    textView.Highlight("0")
    app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if currentChar >= len(words) {
            return event
        }

        if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
            colorMap[currentChar] = "white"

            currentChar -= 1
            if currentChar < 0 { currentChar = 0 }

            textView.Highlight(fmt.Sprintf("%d", currentChar))
            drawText(textView, words, colorMap)
            return event
        }

        if event.Rune() == rune(words[currentChar]) {
            colorMap[currentChar] = "green"
        } else {
            colorMap[currentChar] = "red"
        }

        drawText(textView, words, colorMap)

        currentChar += 1
        if currentChar >= len(words) {
            words, colorMap = newGame(characterPriority)
            textView.Highlight("0")
            currentChar = 0
            drawText(textView, words, colorMap)
        }

        textView.Highlight(fmt.Sprintf("%d", currentChar))

        return event
    })

    textView.SetBorder(true)
    if err := app.SetRoot(textView, true).SetFocus(textView).Run(); err != nil {
        panic(err)
    }
}


func newGame(characterPriority []dictionary.CharacterPriority) (string, []string){
    numChars := 5
    usedCharacters := characterPriority[:numChars]
    priorityCharacter := usedCharacters[0]
    wordCount := 10

    words := ""
    for i := 0; i < wordCount; i++ {
        words += fmt.Sprintf("%s ", generateWord(usedCharacters, priorityCharacter))
    }

    colorMap := make([]string, len(words))
    for i := 0; i < len(words); i++ {
        colorMap[i] = "white"
    }

    return words, colorMap
}


func drawText(textView *tview.TextView, words string, colorMap []string) {
    textView.Clear()

    for i, char := range words {
        fmt.Fprintf(textView, `["%d"][%s]%c[""]`, i, colorMap[i], char)
    }        
} 


func generateWord(chars []dictionary.CharacterPriority, priorityCharacter dictionary.CharacterPriority) string {
    length := rand.Intn(3) + 3
    priorityPosition := rand.Intn(length)

    word := ""
    for i := 0; i < length; i++ {
        if i == priorityPosition {
            word += string(priorityCharacter.Character)
            continue
        }

        word += string(chars[rand.Intn(len(chars))].Character)
    }

    return word
}
