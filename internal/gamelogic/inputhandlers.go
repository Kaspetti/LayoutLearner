package gamelogic

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

// gameInputHandler handles the input from the user when the game is running.
// It checks if the user inputs the correct character according to
// the current character index. When the user reaches the end of the words
// it signals to change the current input capture function to endScreenLogic.
func gameInputHandler(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
        graphicsCtx.App.Stop()
    }

    if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
        graphicsCtx.MainColorMap[gameCtx.CurrentCharIndex] = "white"

        gameCtx.CurrentCharIndex -= 1
        if gameCtx.CurrentCharIndex < 0 { gameCtx.CurrentCharIndex = 0 }

        graphicsCtx.MainTextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))
        graphicsCtx.DrawText(gameCtx.Words, gameCtx.PriorityCharacter, gameCtx.CurrentChars, gameCtx.CharacterAccuracies)
        return event
    }

    if event.Rune() == rune(gameCtx.Words[gameCtx.CurrentCharIndex]) {
        updateAccuracy(rune(gameCtx.Words[gameCtx.CurrentCharIndex]), true)

        if !gameCtx.Started {
            gameCtx.Started = true
            gameCtx.StartTimeCharacter = time.Now().UnixMilli()
        } else {
            ca := gameCtx.CharacterAccuracies[rune(gameCtx.Words[gameCtx.CurrentCharIndex])]
            ca.TotalTime += time.Now().UnixMilli() - gameCtx.StartTimeCharacter
            ca.AverageTime = ca.TotalTime / ca.Attempts

            gameCtx.CharacterAccuracies[rune(gameCtx.Words[gameCtx.CurrentCharIndex])] = ca
        }

        graphicsCtx.MainColorMap[gameCtx.CurrentCharIndex] = "blue"
        gameCtx.Correct += 1
    } else {
        updateAccuracy(rune(gameCtx.Words[gameCtx.CurrentCharIndex]), false)
        graphicsCtx.MainColorMap[gameCtx.CurrentCharIndex] = "red"
        gameCtx.Incorrect += 1
    }

    graphicsCtx.DrawText(gameCtx.Words, gameCtx.PriorityCharacter, gameCtx.CurrentChars, gameCtx.CharacterAccuracies)

    gameCtx.CurrentCharIndex += 1
    if gameCtx.CurrentCharIndex >= len(gameCtx.Words) - 1 {
        graphicsCtx.ShowEndScreen(float64(gameCtx.Correct), float64(gameCtx.Incorrect))
        inputCaptureChangeChan <- endScreenInputHandler
        return event
    }

    graphicsCtx.MainTextView.Highlight(fmt.Sprintf("%d", gameCtx.CurrentCharIndex))
    gameCtx.StartTimeCharacter = time.Now().UnixMilli()

    return event
}


// endScreenInputHandler handles the player input on the end screen.
// From here the player is able to either start a new game with the
// <Enter> key or stop the game using <Escape>. If <Enter> is pressed
// the game context will be reset and the input capture function will
// transition to gameLogic
func endScreenInputHandler(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        newGame()
        graphicsCtx.MainTextView.Highlight("0")
        graphicsCtx.DrawText(gameCtx.Words, gameCtx.PriorityCharacter, gameCtx.CurrentChars, gameCtx.CharacterAccuracies)

        inputCaptureChangeChan <- gameInputHandler
        return nil
    } else if event.Key() == tcell.KeyEscape {
        graphicsCtx.App.Stop()
    }

    return event
}
