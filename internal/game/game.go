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
    Words               string                  // The words of the current game
    ColorMap            []string                // The ColorMap for the characters. The colors of each character is a word representing its the color at that index.
    CurrentCharIndex    int                     // The index of the character currently in play
    CharacterPriority   []rune                  // Slice of all characters in the dictionary sorted by priority
    Correct             int                     // The amount of correctly written characters this round
    Incorrect           int                     // The amount of incorrently written characters this round
    NumChars            int                     // The number of characters from the CharacterPriority to user when generating wwords
    MaxWordLength       int                     // The maximum length of words generated
    WordCount           int                     // Amount of words generated per round
}

type Graphics struct {
    App             *tview.Application          // The tview application for rendering to the terminal
    MainTextView    *tview.TextView             // The main text view where the game takes place
    InfoTextView    *tview.TextView             // An information text view to the right of the main text view
    MainFlex        *tview.Flex                 // The main tview flex box containing all other elements
}

var gameCtx     GameContext
var graphics    Graphics

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
        return err
    }

    gameCtx = GameContext{
        CharacterPriority: characterPriority,
        NumChars: 5,
        MaxWordLength: 5,
        WordCount: 10,
    }
    gameCtx.newGame()

    graphics = Graphics{
        App: tview.NewApplication(),
        MainTextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        InfoTextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        MainFlex: tview.NewFlex(),
    }
    graphics.MainFlex.
        AddItem(graphics.MainTextView, 0, 1, true).
        AddItem(graphics.InfoTextView, 20, 1, false)

    drawText()

    graphics.MainTextView.Highlight("0")
    graphics.App.SetInputCapture(gameLogic)

    // Sets up the goroutine for handling switching of input capture functions
    go func() {
        for {
            changeFunc := <-inputCaptureChangeChan
            graphics.App.QueueUpdate(func() {
                graphics.App.SetInputCapture(changeFunc)
            })
        }
    }()

    graphics.MainTextView.SetBorder(true)
    graphics.InfoTextView.SetBorder(true)
    if err := graphics.App.SetRoot(graphics.MainFlex, true).SetFocus(graphics.MainFlex).Run(); err != nil {
        return err
    }

    return nil
}


// newGame resets the game gontext by generating new words from the 
// character priority and resetting the other fields to their original value.
func (gc *GameContext) newGame() {
    charactersInUse := gc.CharacterPriority[:gc.NumChars]
    priorityCharacter := charactersInUse[0]

    words := ""
    for i := 0; i < gc.WordCount; i++ {
        words += fmt.Sprintf("%s ", dictionary.GenerateWord(charactersInUse, priorityCharacter, gc.MaxWordLength))
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
        graphics.App.Stop()
    }

    if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
        gameCtx.ColorMap[gameCtx.CurrentCharIndex] = "white"

        gameCtx.CurrentCharIndex -= 1
        if gameCtx.CurrentCharIndex < 0 { gameCtx.CurrentCharIndex = 0 }

        graphics.MainTextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))
        drawText()
        return event
    }

    if event.Rune() == rune(gameCtx.Words[gameCtx.CurrentCharIndex]) {
        gameCtx.ColorMap[gameCtx.CurrentCharIndex] = "green"
        gameCtx.Correct += 1
    } else {
        gameCtx.ColorMap[gameCtx.CurrentCharIndex] = "red"
        gameCtx.Incorrect += 1
    }

    drawText()

    gameCtx.CurrentCharIndex += 1
    if gameCtx.CurrentCharIndex >= len(gameCtx.Words) - 1 {
        showEndScreen()
        inputCaptureChangeChan <- endScreenLogic
        return event
    }

    graphics.MainTextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))

    return event

}


// showEndScreen prints the end screen for the game, providing the user 
// with information about their accuracy.
func showEndScreen() {
    graphics.MainTextView.Clear()
    accuracy := float64(gameCtx.Correct * 100) / float64(gameCtx.Correct + gameCtx.Incorrect)

    fmt.Fprintf(graphics.MainTextView, "[white]Your accuracy was: %.2f\n[yellow]Press enter to continue...\n[red]Press escape to exit...", accuracy)
}


// endScreenLogic handles the player input on the end screen.
// From here the player is able to either start a new game with the
// <Enter> key or stop the game using <Escape>. If <Enter> is pressed
// the game context will be reset and the input capture function will
// transition to gameLogic
func endScreenLogic(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        gameCtx.newGame()
        graphics.MainTextView.Highlight("0")
        drawText()

        inputCaptureChangeChan <- gameLogic
        return nil
    } else if event.Key() == tcell.KeyEscape {
        graphics.App.Stop()
    }

    return event
}


// drawText draws the words to the textView giving each character the colors
// by index listed in the given color map.
func drawText() {
    graphics.MainTextView.Clear()

    for i, char := range gameCtx.Words {
        fmt.Fprintf(graphics.MainTextView, `["%d"][%s]%c[""]`, i, gameCtx.ColorMap[i], char)
    }        
} 
