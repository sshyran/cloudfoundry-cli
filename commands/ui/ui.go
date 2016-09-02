// package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
package ui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/fatih/color"

	"code.cloudfoundry.org/cli/utils/config"
	"github.com/nicksnyder/go-i18n/i18n"
)

const (
	red   color.Attribute = color.FgRed
	green                 = color.FgGreen
	// yellow                         = color.FgYellow
	// magenta                        = color.FgMagenta
	cyan = color.FgCyan
	// grey                           = color.FgWhite
	defaultFgColor = 38
)

//go:generate counterfeiter . Config

// Config is the UI configuration
type Config interface {
	// ColorEnabled enables or disabled color
	ColorEnabled() config.ColorSetting

	// Locale is the language to translate the output to
	Locale() string
}

// UI is interface to interact with the user
type UI struct {
	// Out is the output buffer
	Out io.Writer

	// Err is the error buffer
	Err io.Writer

	colorEnabled config.ColorSetting

	translate i18n.TranslateFunc
}

// NewUI will return a UI object where Out is set to STDOUT
func NewUI(c Config) (UI, error) {
	translateFunc, err := GetTranslationFunc(c)
	if err != nil {
		return UI{}, err
	}

	return UI{
		Out:          color.Output,
		Err:          os.Stderr,
		colorEnabled: c.ColorEnabled(),
		translate:    translateFunc,
	}, nil
}

// NewTestUI will return a UI object where Out and Err are customizable
func NewTestUI(out io.Writer, err io.Writer) UI {
	return UI{
		Out:          out,
		Err:          err,
		colorEnabled: config.ColorDisbled,
		translate:    i18n.TranslateFunc(func(s string, _ ...interface{}) string { return s }),
	}
}

// DisplayTable presents a two dimentional array of strings as a table to UI.Out
func (ui UI) DisplayTable(prefix string, table [][]string) {
	tw := tabwriter.NewWriter(ui.Out, 0, 1, 4, ' ', 0)

	for _, row := range table {
		fmt.Fprint(tw, prefix)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	tw.Flush()
}

// DisplayText combines the formattedString template with the key maps and then
// outputs it to the UI.Out file. The maps are merged in a way that the last
// one takes precidence over the first. Prior to outputting the
// formattedString, it is run through the an internationalization function to
// translate it to a pre-configured langauge.
func (ui UI) DisplayText(formattedString string, keys ...map[string]interface{}) {
	ui.outputTextToBuffer(ui.Out, formattedString, true, keys...)
}

// DisplayTextWithKeyTranslations merges keys together (similar to
// DisplayText), translates the keys listed in keysToTranslate, and then passes
// these values to DisplayText.
func (ui UI) DisplayTextWithKeyTranslations(formattedString string, keysToTranslate []string, keys ...map[string]interface{}) {
	mergedMap := ui.mergeMap(keys)
	for _, key := range keysToTranslate {
		mergedMap[key] = ui.translate(mergedMap[key].(string))
	}
	ui.outputTextToBuffer(ui.Out, formattedString, true, mergedMap)
}

// DisplayNewline outputs a newline to UI.Out.
func (ui UI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayPair outputs the "attribute: formattedString" pair to UI.Out. keys are merged
// together and then applied to the translation of formattedString, while attribute is
// translated directly.
func (ui UI) DisplayPair(attribute string, formattedString string, keys ...map[string]interface{}) {
	mergedMap := ui.mergeMap(keys)
	translatedFormatString := ui.translate(formattedString, mergedMap)

	formattedTemplate := template.Must(template.New("Display Text").Parse(translatedFormatString))
	var buffer bytes.Buffer
	formattedTemplate.Execute(&buffer, mergedMap)

	fmt.Fprintf(ui.Out, "%s: %s\n", ui.translate(attribute), buffer.String())
}

// DisplayHelpHeader translates and then bolds the help header. Sends output to
// UI.Out
func (ui UI) DisplayHelpHeader(text string) {
	fmt.Fprintf(ui.Out, ui.colorize(ui.translate(text), defaultFgColor, true))
	ui.DisplayNewline()
}

// DisplayHeaderFlavorText translates text and colorizes the keys after merging
// the maps. The color used is cyan, and the text is outputted to UI.Out.
func (ui UI) DisplayHeaderFlavorText(text string, keys ...map[string]interface{}) {
	mergedMap := ui.mergeMap(keys)
	for key, value := range mergedMap {
		mergedMap[key] = ui.colorize(fmt.Sprint(value), cyan, true)
	}
	ui.outputTextToBuffer(ui.Out, text, true, mergedMap)
}

// DisplayOK will output a green translated "OK" message to UI.Out.
func (ui UI) DisplayOK() {
	translatedFormatString := ui.translate("OK", nil)
	fmt.Fprintf(ui.Out, ui.colorize(translatedFormatString, green, true))
	ui.DisplayNewline()
}

// DisplayFailed will output a red translated "FAILED" message to UI.Out.
func (ui UI) DisplayFailed() {
	translatedFormatString := ui.translate("FAILED", nil)
	fmt.Fprintf(ui.Out, ui.colorize(translatedFormatString, red, true))
	ui.DisplayNewline()
}

// DisplayErrorMessage combines the err template with the key maps and then
// outputs it to the UI.Err file. The maps are merged in a way that the last
// one takes precidence over the first. Prior to outputting the err, it is run
// through the an internationalization function to translate it to a
// pre-configured langauge.
func (ui UI) DisplayErrorMessage(err string, keys ...map[string]interface{}) {
	ui.outputTextToBuffer(ui.Err, err, true, keys...)
}

func (ui UI) mergeMap(maps []map[string]interface{}) map[string]interface{} {
	if len(maps) == 1 {
		return maps[0]
	}

	main := map[string]interface{}{}

	for _, minor := range maps {
		for key, value := range minor {
			main[key] = value
		}
	}

	return main
}

func (ui UI) colorize(message string, textColor color.Attribute, bold bool) string {
	colorPrinter := color.New(textColor)
	switch ui.colorEnabled {
	case config.ColorEnabled:
		colorPrinter.EnableColor()
	case config.ColorDisbled:
		colorPrinter.DisableColor()
	}

	if bold {
		colorPrinter = colorPrinter.Add(color.Bold)
	}
	f := colorPrinter.SprintFunc()
	return f(message)
}

func (ui UI) outputTextToBuffer(writer io.Writer, formattedString string, includeNewline bool, keys ...map[string]interface{}) {
	mergedMap := ui.mergeMap(keys)
	translatedFormatString := ui.translate(formattedString, mergedMap)
	formattedTemplate := template.Must(template.New("Display Text").Parse(translatedFormatString))
	formattedTemplate.Execute(writer, mergedMap)
	if includeNewline {
		fmt.Fprintf(writer, "\n")
	}
}
