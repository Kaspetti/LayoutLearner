// Package shared contains interfaces and types which are shared among multiple 
// packages
package shared


// CharacterAccuracy stores information about the accuracy and time spent
// on a specific character
type CharacterAccuracy struct {
    Correct     int64       `json:"correct"`        // The amount of correct guesses of a character
    Attempts    int64       `json:"attempts"`       // The amount of attempts for each character
    Accuracy    float64     `json:"accuracy"`       // The accuracy of a character (Correct / Attempts)
    TotalTime   int64       `json:"totalTime"`      // The total time spent on all attempts in milliseconds
    AverageTime int64       `json:"averageTime"`    // The average time spent per attempt in milliseconds
    Score       float64     `json:"score"`          // The total score of the character considering accuracy and time
}
