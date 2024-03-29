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


// GetCharacterPriority returns a list of character priorities for each character in a
// dictionary given the path of the dictionary file.
func GetCharacterPriority(dictionaryPath string) ([]rune, error) {
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

    characters := make([]rune, len(characterOccurences)) 
    i := 0
    for char := range characterOccurences {
        characters[i] = char
        i += 1
    }

    sort.Slice(characters, func(i, j int) bool {
        return characterOccurences[characters[i]] > characterOccurences[characters[j]]
    })

    return characters, nil
}


// GenerateWord generates a random word using the characters provided. The caller may choose the
// length of the word and a priority character. The priority character is guaranteed to be within
// the word. 
func GenerateWord(chars []rune, priorityCharacter rune, minLength, maxLength int) string {
    length := rand.Intn(maxLength-minLength) + minLength
    priorityPosition := rand.Intn(length)

    charsUsed := make(map[rune]int)
    for _, char := range chars {
        charsUsed[char] = 0
    }

    var previousCharacter rune
    charInARow := 0

    word := ""
    for i := 0; i < length; i++ {
        if i == priorityPosition && previousCharacter != priorityCharacter {
            word += string(priorityCharacter)
            charsUsed[priorityCharacter] += 1
            continue
        }

        var excludeChar rune
        if charInARow == 2 {
            excludeChar = previousCharacter
        }
        char := getRandomCharacter(chars, charsUsed, length/2, excludeChar)

        // Makes sure the loop breaks if there are no characters possible to use
        if char == ' ' {
            break
        }

        word += string(char)
        charsUsed[char] += 1

        if char == previousCharacter {
            charInARow += 1
        } else {
            charInARow = 1
        }

        previousCharacter = char
    }

    return word
}


// getRandomCharacter gets a random character from chars which has not been used more
// than maxUsage.
func getRandomCharacter(chars []rune, charsUsed map[rune]int, maxUsage int, exclude rune) rune {
    availableChars := make([]rune, 0)
    for _, char := range chars {
        if char == exclude {
            continue
        }

        if charsUsed[char] < maxUsage {
            availableChars = append(availableChars, char)
        }
    }

    if len(availableChars) > 0 {
        return availableChars[rand.Intn(len(availableChars))]
    } else {
        return ' '
    }
}


// GetWordsFromChars gets "amount" of words from the dictionary passed to it which use only the characters in "chars",  
// which contain the "priorityChar" and satisfy the min and max length.
func GetWordsFromChars(dictionaryPath string, chars []rune, priorityChar rune, minLength, maxLength, amount int) ([]string, error) {
    f, err := os.Open(dictionaryPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    charsSet := make(map[rune]bool)
    for _, char := range chars {
        charsSet[char] = true
    }

    words := make([]string, 0)

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        priorityFound := false
        invalidChar := false
        word := strings.ToLower(scanner.Text())

        if len(word) > maxLength || len(word) < minLength {
            continue
        }

        for _, char := range word {
            if !charsSet[char] {
                invalidChar = true
                break
            }

            if char == priorityChar {
                priorityFound = true
            }
        }

        if !invalidChar && priorityFound {
            words = append(words, word)
        }
    }

    if len(words) < 4 {
        for i := 0; i < 4 - len(words); i++ {
            words = append(words, GenerateWord(chars, priorityChar, minLength, maxLength))
        }
    }

    selectedWords := make([]string, amount)
    for i := 0; i < amount; i++ {
        selectedWords[i] = words[rand.Intn(len(words))]
    }


    return selectedWords, nil
}
