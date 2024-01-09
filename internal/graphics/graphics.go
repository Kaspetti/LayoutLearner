// Package graphics handles showing the terminal user interface of the game. It uses tview
// to draw to the terminal.
package graphics

import (
	"fmt"
	"image/color"

	"github.com/Kaspetti/LayoutLearner/internal/shared"
	"github.com/rivo/tview"
)

// GraphicsContext stores all elements for showing the TUI
type GraphicsContext struct {
    App                 *tview.Application          // The tview application for rendering to the terminal
    MainTextView        *tview.TextView             // The main text view where the game takes place
    InfoTextView        *tview.TextView             // An information text view to the right of the main text view
    MainFlex            *tview.Flex                 // The main tview flex box containing all other elements
    MainColorMap        []string                    // The color map for the characters. The colors of each character is a word representing its the color at that index.
}


func InitializeGraphics() GraphicsContext {
    graphicsCtx := GraphicsContext{
        App: tview.NewApplication(),
        MainTextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        InfoTextView: tview.NewTextView().SetRegions(true).SetDynamicColors(true),
        MainFlex: tview.NewFlex(),
    }
    graphicsCtx.MainFlex.
        AddItem(graphicsCtx.MainTextView, 0, 1, true).
        AddItem(graphicsCtx.InfoTextView, 31, 1, false)

    graphicsCtx.MainTextView.SetBorder(true)
    graphicsCtx.InfoTextView.SetBorder(true)

    graphicsCtx.MainTextView.Highlight("0")

    return graphicsCtx
}


// DrawText draws the words to the textView giving each character the colors
// by index listed in the given color map.
func (gc *GraphicsContext) DrawText(words string, priorityChar rune, currentChars []rune, characterAccuracies map[rune]shared.CharacterAccuracy) {
    gc.MainTextView.Clear()
    gc.InfoTextView.Clear()

    // Draw the words to the main text view
    for i, char := range words {
        if char == ' ' && i < len(words) - 1{
            fmt.Fprintf(gc.MainTextView, `["%d"][%s][::u] [::-]`, i, gc.MainColorMap[i])
            continue
        }
        fmt.Fprintf(gc.MainTextView, `["%d"][%s]%c[""]`, i, gc.MainColorMap[i], char)
    }        

    // Draw information
    fmt.Fprint(gc.InfoTextView, "[yellow]Accuracy:\n")
    for i, char := range currentChars {
        color := "white"
        if characterAccuracies[char].Accuracy != -1 {
            color = interpolateColor(characterAccuracies[char].Accuracy)
        } 
        fmt.Fprintf(
            gc.InfoTextView,
            `["usedChars"][%s]%c[""]`,
            color,
            char,
        )

        if i < len(currentChars) - 1 {
            fmt.Fprint(gc.InfoTextView, "[white][u]|")
        }
    }

    priortiyColor := "white"
    if characterAccuracies[priorityChar].Accuracy != -1 {
        priortiyColor = interpolateColor(characterAccuracies[priorityChar].Accuracy)
    }
    fmt.Fprintf(gc.InfoTextView, "\n\n[yellow]Priority: [%s][\"usedChars\"]%c[\"\"][white]", priortiyColor, priorityChar)

    fmt.Fprintf(gc.InfoTextView, "\n\n[yellow]Average times:")
    for char, ca := range characterAccuracies {
        if ca.AverageTime >= 1000 {
            averageTime := float64(ca.AverageTime) / 1000.0
            fmt.Fprintf(gc.InfoTextView, "\n%c=%.2fs", char, averageTime)
            continue
        }
        fmt.Fprintf(gc.InfoTextView, "\n%c=%dms", char, ca.AverageTime)
    }


    gc.InfoTextView.Highlight("usedChars")
} 


// showEndScreen prints the end screen for the game, providing the user 
// with information about their accuracy.
func (gc *GraphicsContext) ShowEndScreen(correct, incorrect float64) {
    gc.MainTextView.Clear()
    accuracy := (correct * 100) / (correct + incorrect)

    fmt.Fprintf(gc.MainTextView, "[white]Your accuracy was: %.2f\n[yellow]Press enter to continue...\n[red]Press escape to exit...", accuracy)
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
