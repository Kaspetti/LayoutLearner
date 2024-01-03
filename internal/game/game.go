// Package game contains functionality for handling the game logic.
// The game logic includes drawing the game to the terminal and handling
// any input for each stage of the game.
package game

import (
	"fmt"

	"github.com/Kaspetti/LayoutLearner/internal/dictionary"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// GameContext stores information of the game.
type GameContext struct {
    Words               string                              // The words of the current game
    ColorMap            []string                            // The ColorMap for the characters. The colors of each character is a word representing its the color at that index.
    CurrentCharIndex    int                                 // The index of the character currently in play
    CharacterPriority   []rune                              // Slice of all characters in the dictionary sorted by priority
    App                 *tview.Application                  // The tview application for rendering to the terminal
    TextView            *tview.TextView                     // The tview textview element for showing the text
    Correct             int                                 // The amount of correctly written characters this round
    Incorrect           int                                 // The amount of incorrently written characters this round
}

var gameCtx GameContext

// Channel for handling changes in the input capture. This is handled by a channel 
// and a goroutine as changing the input capture function does not work as 
// expected when changed from within the current input capture function.
var inputCaptureChangeChan = make(chan func(*tcell.EventKey) *tcell.EventKey)


// StartGame starts the game. It gets the character priorities of the 
// dictionary in use and creates the tview application and textview.
// It then creates a fresh game context and starts the goroutine for
// handling input capture function changes.
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


// newGame resets the game gontext by generating new words from the 
// character priority and resetting the other fields to their original value.
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


// gameLogic handles the input from the user when the game is running.
// It checks if the user inputs the correct character according to 
// the current character index. When the user reaches the end of the words
// it signals to change the current input capture function to endScreenLogic.
func gameLogic(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
        gameCtx.App.Stop()
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


// showEndScreen prints the end screen for the game, providing the user 
// with information about their accuracy.
func showEndScreen() {
    gameCtx.TextView.Clear()
    accuracy := float64(gameCtx.Correct * 100) / float64(gameCtx.Correct + gameCtx.Incorrect)

    fmt.Fprintf(gameCtx.TextView, "[white]Your accuracy was: %.2f\n[yellow]Press enter to continue...\n[red]Press escape to exit...", accuracy)
}


// endScreenLogic handles the player input on the end screen.
// From here the player is able to either start a new game with the
// <Enter> key or stop the game using <Escape>. If <Enter> is pressed
// the game context will be reset and the input capture function will
// transition to gameLogic
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


// drawText draws the words to the textView giving each character the colors
// by index listed in the given color map.
func drawText(textView *tview.TextView, words string, colorMap []string) {
    textView.Clear()

    for i, char := range words {
        fmt.Fprintf(textView, `["%d"][%s]%c[""]`, i, colorMap[i], char)
    }        
} 
