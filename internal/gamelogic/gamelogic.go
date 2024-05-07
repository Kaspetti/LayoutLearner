package gamelogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

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
    Started             bool                                // Started becomes true the moment the player hits a button
    StartTimeCharacter  int64                               // The time when the current character went into play in milliseconds since unix
    Settings            GameSettings                        // The settings for the game
}

// GameSettings stores the settings for the game. AccuracyWeight and TimeWeight should add up to 1.0
type GameSettings struct {
    NumChars            int                                 // The number of characters to use from CharacterPriorities
    MaxWordLength       int                                 // The max word length (inclusive)
    MinWordLength       int                                 // The min word length (inclusive)
    WordCount           int                                 // The amount of words to include in the lesson
    TargetCPM           int                                 // The target "characters per minute" used for scoring
    AccuracyWeight      float64                             // The weight at which accuracy affects the final score
    TimeWeight          float64                             // The weight at which speed affects the final score
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

    var charAccuracies map[rune]shared.CharacterAccuracy
    if _, err := os.Stat("accuracies"); errors.Is(err, os.ErrNotExist) {
        charAccuracies = make(map[rune]shared.CharacterAccuracy)
    } else {
        saveData, err := os.ReadFile("accuracies")
        if err != nil {
            return err
        }

        if err := json.Unmarshal(saveData, &charAccuracies); err != nil {
            return err
        }
    }

    gameCtx = GameContext{
        CharacterPriorities: characterPriority,
        CharacterAccuracies: charAccuracies,
        
        Settings: GameSettings{
            NumChars: 5,
            MinWordLength: 3,
            MaxWordLength: 5,
            WordCount: 10,
            TargetCPM: 250,
            TimeWeight: 0.5,
            AccuracyWeight: 0.5,
        },
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
    gameCtx.CurrentChars = gameCtx.CharacterPriorities[:gameCtx.Settings.NumChars]
    gameCtx.PriorityCharacter = getPriorityCharacter()

    wordsList, err := dictionary.GetWordsFromChars(
        "resources/words.txt", 
        gameCtx.CurrentChars, 
        gameCtx.PriorityCharacter, 
        gameCtx.Settings.MinWordLength, 
        gameCtx.Settings.MaxWordLength, 
        gameCtx.Settings.WordCount,
    )
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

    if err := SaveCharacterAccuracies(); err != nil {
        log.Fatalln(err)
    }

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
    targetSpeedMs := 60000 / gameCtx.Settings.TargetCPM
    lowerBound := targetSpeedMs / 2
    speed := ca.AverageTime
    if speed < int64(lowerBound) {
        speed = int64(lowerBound)
    }
    speedScore := 1 - (float64(speed - int64(lowerBound)) / float64(targetSpeedMs - lowerBound))

    ca.Score = (ca.Accuracy * gameCtx.Settings.AccuracyWeight) + (speedScore * gameCtx.Settings.TimeWeight)

    // Update or add the character accuracy in the map
    gameCtx.CharacterAccuracies[char] = ca
}


func getPriorityCharacter() rune {
    least := 1.0
    priorityChar := gameCtx.CurrentChars[0]

    for char, ca := range gameCtx.CharacterAccuracies {
        if char == ' ' {
            continue
        }

        if ca.Score < least {
            least = ca.Score
            priorityChar = char 
        }
    }

    return priorityChar
}


func SaveCharacterAccuracies() error {
    b, err := json.Marshal(gameCtx.CharacterAccuracies)
    if err != nil {
        return err
    }   

    file, errs := os.Create("accuracies")
    if errs != nil {
        return err
    }
    defer file.Close()

    _, err = file.WriteString(string(b))
    if err != nil {
        return err
    }
    
    return nil
}


func deleteSave() error {
    if err := os.Remove("accuracies"); err != nil {
        return err
    }

    gameCtx.CharacterAccuracies = make(map[rune]shared.CharacterAccuracy)

    return nil
}
