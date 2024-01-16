package gamelogic

import (
	"fmt"

	"github.com/Kaspetti/LayoutLearner/internal/dictionary"
	"github.com/Kaspetti/LayoutLearner/internal/graphics"
	"github.com/Kaspetti/LayoutLearner/internal/shared"
	"github.com/gdamore/tcell/v2"
)

// GameContext stores information of the game.
type GameContext struct {
    Words               string                              // The words of the current game
    CurrentCharIndex    int                                 // The index of the character currently in play
    CharacterPriorities []rune                              // Slice of all characters in the dictionary sorted by priority
    PriorityCharacter   rune                                // The priority character to include in each word
    CurrentChars        []rune                              // Slice of the currently used characters in each lesson
    CharacterAccuracies map[rune]shared.CharacterAccuracy   // The accuracy the user has with each character
    Correct             int                                 // The amount of correctly written characters this round
    Incorrect           int                                 // The amount of incorrently written characters this round
    NumChars            int                                 // The number of characters from the CharacterPriority to user when generating wwords
    MaxWordLength       int                                 // The maximum length of words generated
    MinWordLength       int                                 // The minimum length of the words generated
    WordCount           int                                 // Amount of words generated per round
    Started             bool                                // Started becomes true the moment the player hits a button
    StartTimeCharacter  int64                               // The time when the current character went into play in milliseconds since unix
    TargetCPM           int64                               // The target "characters per minute" for the user. Used for calculating speed score
    AccuracyWeight      float64                             // The weight of the character accuracy when calculating the total score
    TimeWeight          float64                             // The weight of the time score in calcualting the total score
}


var gameCtx     GameContext
var graphicsCtx graphics.GraphicsContext


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
        CharacterPriorities: characterPriority,
        NumChars: 5,
        MinWordLength: 3,
        MaxWordLength: 5,
        WordCount: 10,
        CharacterAccuracies: make(map[rune]shared.CharacterAccuracy),
        TargetCPM: 250,
        TimeWeight: 0.5,
        AccuracyWeight: 0.5,
    }

    graphicsCtx = graphics.InitializeGraphics()
    graphicsCtx.App.SetInputCapture(gameInputHandler)

    // Sets up the goroutine for handling switching of input capture functions
    go func() {
        for {
            changeFunc := <-inputCaptureChangeChan
            graphicsCtx.App.QueueUpdate(func() {
                graphicsCtx.App.SetInputCapture(changeFunc)
            })
        }
    }()

    newGame()
    if err := graphicsCtx.App.SetRoot(graphicsCtx.MainFlex, true).Run(); err != nil {
        return err
    }

    return nil
}


// newGame resets the game gontext by generating new words from the 
// character priority and resetting the other fields to their original value.
func newGame() {
    gameCtx.CurrentChars = gameCtx.CharacterPriorities[:gameCtx.NumChars]
    gameCtx.PriorityCharacter = getPriorityCharacter()

    wordsList, err := dictionary.GetWordsFromChars("resources/words.txt", gameCtx.CurrentChars, gameCtx.PriorityCharacter, gameCtx.MinWordLength, gameCtx.MaxWordLength, gameCtx.WordCount)
    if err != nil {
        graphicsCtx.ShowErrorScreen("generating new words", err)
        inputCaptureChangeChan <- endScreenInputHandler 
        return
    }

    words := ""
    for _, word := range wordsList {
        words += fmt.Sprintf("%s ", word)
    }

    colorMap := make([]string, len(words))
    for i := 0; i < len(words); i++ {
        colorMap[i] = "white"
    }

    for _, char := range gameCtx.CurrentChars {
        if _, ok := gameCtx.CharacterAccuracies[char]; !ok {
            gameCtx.CharacterAccuracies[char] = shared.CharacterAccuracy {
                Attempts: 0,
                Correct: 0,
                Score: -1,
            }
        }
    }

    gameCtx.Words = words
    graphicsCtx.MainColorMap = colorMap
    gameCtx.CurrentCharIndex = 0
    gameCtx.Correct = 0
    gameCtx.Incorrect = 0
    gameCtx.Started = false

    graphicsCtx.MainTextView.Highlight("0")
    graphicsCtx.DrawText(gameCtx.Words, gameCtx.PriorityCharacter, gameCtx.CurrentChars, gameCtx.CharacterAccuracies)

    inputCaptureChangeChan <- gameInputHandler
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

    // Get the target speed per character in ms depending on the TargetCPM
    targetSpeedMs := 60000 / gameCtx.TargetCPM
    lowerBound := targetSpeedMs / 2
    speed := ca.AverageTime
    if speed < lowerBound {
        speed = lowerBound
    }
    speedScore := 1 - (speed - lowerBound) / (targetSpeedMs - lowerBound)

    ca.Score = (ca.Accuracy * gameCtx.AccuracyWeight) + (float64(speedScore) * gameCtx.TimeWeight)

    // Update or add the character accuracy in the map
    gameCtx.CharacterAccuracies[char] = ca
}


func getPriorityCharacter() rune {
    least := 1.0
    priorityChar := gameCtx.CurrentChars[0]

    for char, ca := range gameCtx.CharacterAccuracies {
        if ca.Score < least {
            least = ca.Score
            priorityChar = char 
        }
    }

    return priorityChar
}
