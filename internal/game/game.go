// Package game contains functionality for handling the game logic.
// The game logic includes drawing the game to the terminal and handling
// any input for each stage of the game.
package game

import (
	"fmt"
	"image/color"

	"github.com/Kaspetti/LayoutLearner/internal/dictionary"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CharacterAccuracy struct {
    Correct     float64
    Attempts    float64
    Accuracy    float64
}


// GameContext stores information of the game.
type GameContext struct {
    Words               string                      // The words of the current game
    CurrentCharIndex    int                         // The index of the character currently in play
    CharacterPriority   []rune                      // Slice of all characters in the dictionary sorted by priority
    CurrentChars        []rune                      // Slice of the currently used characters in each lesson
    CharacterAccuracies map[rune]CharacterAccuracy  // The accuracy the user has with each character
    Correct             int                         // The amount of correctly written characters this round
    Incorrect           int                         // The amount of incorrently written characters this round
    NumChars            int                         // The number of characters from the CharacterPriority to user when generating wwords
    MaxWordLength       int                         // The maximum length of words generated
    WordCount           int                         // Amount of words generated per round
}


// Graphics stores all elements for showing the TUI
type Graphics struct {
    App                 *tview.Application          // The tview application for rendering to the terminal
    MainTextView        *tview.TextView             // The main text view where the game takes place
    InfoTextView        *tview.TextView             // An information text view to the right of the main text view
    MainFlex            *tview.Flex                 // The main tview flex box containing all other elements
    MainColorMap        []string                    // The color map for the characters. The colors of each character is a word representing its the color at that index.
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
        NumChars: 15,
        MaxWordLength: 5,
        WordCount: 10,
        CharacterAccuracies: make(map[rune]CharacterAccuracy),
    }
    graphics = Graphics{
        App: tview.NewApplication(),
        MainTextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        InfoTextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        MainFlex: tview.NewFlex(),
    }
    graphics.MainFlex.
        AddItem(graphics.MainTextView, 0, 1, true).
        AddItem(graphics.InfoTextView, 20, 1, false)


    newGame()
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
    if err := graphics.App.SetRoot(graphics.MainFlex, true).Run(); err != nil {
        return err
    }

    return nil
}


// newGame resets the game gontext by generating new words from the 
// character priority and resetting the other fields to their original value.
func newGame() {
    gameCtx.CurrentChars = gameCtx.CharacterPriority[:gameCtx.NumChars]
    priorityCharacter := gameCtx.CurrentChars[0]

    words := ""
    for i := 0; i < gameCtx.WordCount; i++ {
        words += fmt.Sprintf("%s ", dictionary.GenerateWord(gameCtx.CurrentChars, priorityCharacter, gameCtx.MaxWordLength))
    }

    colorMap := make([]string, len(words))
    for i := 0; i < len(words); i++ {
        colorMap[i] = "white"
    }

    for _, char := range gameCtx.CurrentChars {
        if _, ok := gameCtx.CharacterAccuracies[char]; !ok {
            gameCtx.CharacterAccuracies[char] = CharacterAccuracy{
                Attempts: 0,
                Correct: 0,
                Accuracy: -1,
            }
        }
    }

    gameCtx.Words = words
    graphics.MainColorMap = colorMap
    gameCtx.CurrentCharIndex = 0
    gameCtx.Correct = 0
    gameCtx.Incorrect = 0
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
        graphics.MainColorMap[gameCtx.CurrentCharIndex] = "white"

        gameCtx.CurrentCharIndex -= 1
        if gameCtx.CurrentCharIndex < 0 { gameCtx.CurrentCharIndex = 0 }

        graphics.MainTextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))
        drawText()
        return event
    }

    if event.Rune() == rune(gameCtx.Words[gameCtx.CurrentCharIndex]) {
        graphics.MainColorMap[gameCtx.CurrentCharIndex] = "blue"
        gameCtx.Correct += 1

        updateAccuracy(rune(gameCtx.Words[gameCtx.CurrentCharIndex]), true)
    } else {
        graphics.MainColorMap[gameCtx.CurrentCharIndex] = "red"
        gameCtx.Incorrect += 1

        updateAccuracy(rune(gameCtx.Words[gameCtx.CurrentCharIndex]), false)
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


// updateAccuracy updates the accuracy of a rune given if the attempt
// was a success or not
func updateAccuracy(char rune, success bool) {
    ca := gameCtx.CharacterAccuracies[char]
    ca.Attempts++
    
    if success {
        ca.Correct++
    }

    // Make sure to handle the case where Attempts is zero to avoid division by zero
    if ca.Attempts > 0 {
        ca.Accuracy = float64(ca.Correct) / float64(ca.Attempts)
    } else {
        ca.Accuracy = 0.0
    }

    // Update or add the character accuracy in the map
    gameCtx.CharacterAccuracies[char] = ca
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
        newGame()
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
    graphics.InfoTextView.Clear()

    // Draw the words to the main text view
    for i, char := range gameCtx.Words {
        fmt.Fprintf(graphics.MainTextView, `["%d"][%s]%c[""]`, i, graphics.MainColorMap[i], char)
    }        

    // Draw information
    for i, char := range gameCtx.CurrentChars {
        color := "white"
        if gameCtx.CharacterAccuracies[char].Accuracy != -1 {
            color = interpolateColor(gameCtx.CharacterAccuracies[char].Accuracy)
        } 
        fmt.Fprintf(
            graphics.InfoTextView,
            `["usedChars"][%s]%c[""]`,
            color,
            char,
        )

        if i < len(gameCtx.CurrentChars) - 1 {
            fmt.Fprint(graphics.InfoTextView, "[white]|")
        }
    }
    graphics.InfoTextView.Highlight("usedChars")
} 


func interpolateColor(t float64) string {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Define the RGB values for each color stop
	color0 := color.RGBA{231, 72, 86, 255}
	color1 := color.RGBA{249, 241, 165, 255}
	color2 := color.RGBA{59, 120, 255, 255}

	// Interpolate between the colors based on 't'
	var r, g, b uint8
	r = uint8(float64(color0.R)*(1-t) + float64(color1.R)*t)
	g = uint8(float64(color0.G)*(1-t) + float64(color1.G)*t)
	b = uint8(float64(color0.B)*(1-t) + float64(color1.B)*t)

	if t >= 0.5 {
		t = (t - 0.5) * 2 // Scale t for interpolation between color1 and color2
		r = uint8(float64(color1.R)*(1-t) + float64(color2.R)*t)
		g = uint8(float64(color1.G)*(1-t) + float64(color2.G)*t)
		b = uint8(float64(color1.B)*(1-t) + float64(color2.B)*t)
	}

	// Convert the interpolated color to hex format
	colorHex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	return colorHex
}
