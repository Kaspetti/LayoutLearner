// Package shared contains interfaces and types which are shared among multiple 
// packages
package shared


// CharacterAccuracy stores information about the accuracy and time spent
// on a specific character
type CharacterAccuracy struct {
    Correct     int64           // The amount of correct guesses of a character
    Attempts    int64           // The amount of attempts for each character
    Accuracy    float64         // The accuracy of a character (Correct / Attempts)
    TotalTime   int64           // The total time spent on all attempts in milliseconds
    AverageTime int64           // The average time spent per attempt in milliseconds
}
