package dictionary

import (
	"bufio"
	"math/rand"
	"os"
	"sort"
	"strings"
)


type CharacterPriority struct {
    Character   rune
    Priority    float64
}


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
