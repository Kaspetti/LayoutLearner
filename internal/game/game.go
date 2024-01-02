package game

import (
	"fmt"

	"github.com/Kaspetti/LayoutLearner/internal/dictionary"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type GameContext struct {
    Words               string
    ColorMap            []string
    CurrentCharIndex    int
    CharacterPriority   []dictionary.CharacterPriority
    App                 *tview.Application
    TextView            *tview.TextView
    Correct             int
    Incorrect           int
}

var gameCtx GameContext
var inputCaptureChangeChan = make(chan func(*tcell.EventKey) *tcell.EventKey)


func StartGame() error {
    characterPriority, err := dictionary.GetCharacterPriority("resources/words.txt")
    if err != nil {
        panic(err)
    }

    gameCtx = GameContext{
        App: tview.NewApplication(),
        TextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        CharacterPriority: characterPriority,
    }
    gameCtx.newGame()

    drawText(gameCtx.TextView, gameCtx.Words, gameCtx.ColorMap)

    gameCtx.TextView.Highlight("0")
    gameCtx.App.SetInputCapture(gameLogic)

    go func() {
        for {
            changeFunc := <-inputCaptureChangeChan
                gameCtx.App.QueueUpdate(func() {
                    gameCtx.App.SetInputCapture(changeFunc)
                })
        }
    }()

    gameCtx.TextView.SetBorder(true)
    if err := gameCtx.App.SetRoot(gameCtx.TextView, true).SetFocus(gameCtx.TextView).Run(); err != nil {
        return err
    }

    return nil
}


func (gc *GameContext) newGame() {
    numChars := 5
    charactersInUse := gc.CharacterPriority[:numChars]
    priorityCharacter := charactersInUse[0]
    wordCount := 5

    words := ""
    for i := 0; i < wordCount; i++ {
        words += fmt.Sprintf("%s ", dictionary.GenerateWord(charactersInUse, priorityCharacter))
    }

    colorMap := make([]string, len(words))
    for i := 0; i < len(words); i++ {
        colorMap[i] = "white"
    }

    gc.Words = words
    gc.ColorMap = colorMap
    gc.CurrentCharIndex = 0
    gc.Correct = 0
    gc.Incorrect = 0
}


func gameLogic(event *tcell.EventKey) *tcell.EventKey {
    if gameCtx.CurrentCharIndex >= len(gameCtx.Words) {
        return event
    }

    if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
        gameCtx.ColorMap[gameCtx.CurrentCharIndex] = "white"

        gameCtx.CurrentCharIndex -= 1
        if gameCtx.CurrentCharIndex < 0 { gameCtx.CurrentCharIndex = 0 }

        gameCtx.TextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))
        drawText(gameCtx.TextView, gameCtx.Words, gameCtx.ColorMap)
        return event
    }

    if event.Rune() == rune(gameCtx.Words[gameCtx.CurrentCharIndex]) {
        gameCtx.ColorMap[gameCtx.CurrentCharIndex] = "green"
        gameCtx.Correct += 1
    } else {
        gameCtx.ColorMap[gameCtx.CurrentCharIndex] = "red"
        gameCtx.Incorrect += 1
    }

    drawText(gameCtx.TextView, gameCtx.Words, gameCtx.ColorMap)

    gameCtx.CurrentCharIndex += 1
    if gameCtx.CurrentCharIndex >= len(gameCtx.Words) - 1 {
        showEndScreen()
        inputCaptureChangeChan <- endScreenLogic
        return event
    }

    gameCtx.TextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))

    return event

}


func showEndScreen() {
    gameCtx.TextView.Clear()
    accuracy := float64(gameCtx.Correct * 100) / float64(gameCtx.Correct + gameCtx.Incorrect)

    fmt.Fprintf(gameCtx.TextView, "[white]Your accuracy was: %.2f\n[yellow]Press enter to continue...\n[red]Press escape to exit...", accuracy)
}


func endScreenLogic(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        gameCtx.newGame()
        gameCtx.TextView.Highlight("0")
        drawText(gameCtx.TextView, gameCtx.Words, gameCtx.ColorMap)

        inputCaptureChangeChan <- gameLogic
        return nil
    } else if event.Key() == tcell.KeyEscape {
        gameCtx.App.Stop()
    }

    return event
}


func drawText(textView *tview.TextView, words string, colorMap []string) {
    textView.Clear()

    for i, char := range words {
        fmt.Fprintf(textView, `["%d"][%s]%c[""]`, i, colorMap[i], char)
    }        
} 
