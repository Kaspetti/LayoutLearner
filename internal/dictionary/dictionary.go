// Package dictionary contains functionality for any dictonary operations for the 
// layout learner. Included are functions for getting the most used characters given a
// file of words, and also for generating random "words" given som parameters.
package dictionary

import (
	"bufio"
	"math/rand"
	"os"
	"sort"
	"strings"
)


// CharacterPriority contains a character and its priority.
// The priority of a character is the ratio of how often it 
// occurs in a dictionary.
type CharacterPriority struct {
    Character   rune
    Priority    float64
}


// GetCharacterPriority returns a list of character priorities for each character in a 
// dictionary given the path of the dictionary file.
func GetCharacterPriority(dictionaryPath string) ([]CharacterPriority, error) {
    f, err := os.Open(dictionaryPath)
    if err != nil {
        return nil, err
    } 
    defer f.Close()

    scanner := bufio.NewScanner(f)
    characterOccurences := make(map[rune]int)
    totalCharacterCount := 0

    for scanner.Scan() {
        word := strings.ToLower(scanner.Text())
        for _, char := range word {
            if occurence, ok := characterOccurences[char]; ok {
                characterOccurences[char] = occurence + 1
            } else {
                characterOccurences[char] = 1
            }
            totalCharacterCount += 1
        }
    }

    characterPriorities := make([]CharacterPriority, len(characterOccurences)) 
    i := 0
    for char, occurence := range characterOccurences {
        characterPriorities[i] = CharacterPriority{
            Character: char,
            Priority: float64(occurence) / float64(totalCharacterCount),
        } 
        i += 1
    }

    sort.Slice(characterPriorities, func(i, j int) bool {
        return characterPriorities[i].Priority > characterPriorities[j].Priority
    })

    return characterPriorities, nil
}


// GenerateWord generates a random word using the characters provided. The caller may choose the
// length of the word and a priority character. The priority character is guaranteed to be within
// the word. 
func GenerateWord(chars []CharacterPriority, priorityCharacter CharacterPriority) string {
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
