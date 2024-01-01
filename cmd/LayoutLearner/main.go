package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Kaspetti/LayoutLearner/internal/dictionary"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


var words = ""
var colorMap = []string{}
var currentChar = 0
var textView = tview.NewTextView().SetDynamicColors(true).SetRegions(true)
var characterPriority = []dictionary.CharacterPriority{}
var app = tview.NewApplication()

var correct = 0
var incorrect = 0


func main() {
    rand.Seed(time.Now().UTC().UnixNano())

    var err error
    characterPriority, err = dictionary.GetCharacterPriority("resources/words.txt")
    if err != nil {
        panic(err)
    }

    words, colorMap = newGame(characterPriority)

    drawText(textView, words, colorMap)

    textView.Highlight("0")
    app.SetInputCapture(gameLogic)

    textView.SetBorder(true)
    if err := app.SetRoot(textView, true).SetFocus(textView).Run(); err != nil {
        panic(err)
    }
}


func endScreenLogic(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        words, colorMap = newGame(characterPriority)
        currentChar = 0 
        correct = 0
        incorrect = 0
        textView.Highlight("0")
        drawText(textView, words, colorMap)

        textView.SetInputCapture(gameLogic)
        return event
    } else if event.Key() == tcell.KeyEscape {
        app.Stop()
    }

    return event
}


func gameLogic(event *tcell.EventKey) *tcell.EventKey {
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
        correct += 1
    } else {
        colorMap[currentChar] = "red"
        incorrect += 1
    }

    drawText(textView, words, colorMap)

    currentChar += 1
    if currentChar >= len(words) - 1 {
        showEndScreen()

        textView.SetInputCapture(endScreenLogic)
        return event
    }

    textView.Highlight(fmt.Sprintf("%d", currentChar))

    return event

}


func showEndScreen() {
    textView.Clear()
    accuracy := float64(correct * 100) / float64(correct + incorrect)

    fmt.Fprintf(textView, "[white]Your accuracy was: %.2f\n[yellow]Press enter to continue...\n[red]Press escape to exit...", accuracy)
}


func newGame(characterPriority []dictionary.CharacterPriority) (string, []string){
    numChars := 5
    usedCharacters := characterPriority[:numChars]
    priorityCharacter := usedCharacters[0]
    wordCount := 5

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
