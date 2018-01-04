package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

const codeMarker = "##CODE##"

func renderSlide(s slide, index int) string {
	content := s.content
	scanner := bufio.NewScanner(strings.NewReader(content))
	slideMarkup := startSlide(index)
	code := ""

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
			line = strings.TrimPrefix(line, "  ")
			line = strings.TrimPrefix(line, "\t")
			code += line + "\n"
			slideMarkup += codeMarker
		} else if strings.HasPrefix(line, "#") {
			line = strings.TrimPrefix(line, "#")
			line = headline(line)
			slideMarkup += line
		} else {
			if strings.HasPrefix(line, ".") {
				line = strings.TrimPrefix(line, ".")
			} else {
				if strings.Contains(line, "*") {
					charScanner := bufio.NewScanner(strings.NewReader(line))
					line = ""
					charScanner.Split(bufio.ScanRunes)
					emOpen := false
					for charScanner.Scan() {
						c := charScanner.Text()
						if c == "*" && !emOpen {
							c = "<strong>"
							emOpen = true
						} else if c == "*" && emOpen {
							c = "</strong>"
							emOpen = false
						}
						line += c
					}
				}
			}
			slideMarkup += fmt.Sprintf(`
				<p>%s</p>
			`, line)
		}
	}

	markers := strings.Count(slideMarkup, codeMarker)
	slideMarkup = strings.Replace(slideMarkup, codeMarker, "", markers-1)
	if code != "" {
		highlightedCode, cssClasses := getHighlightedMarkup(code, s.code)
		slideMarkup = strings.Replace(slideMarkup, codeMarker, highlightedCode, 1)
		slideMarkup += fmt.Sprintf(`
			<style>%s</style>
		`, cssClasses)
	}

	slideMarkup += endSlide()
	return slideMarkup
}

func headline(txt string) string {
	return fmt.Sprintf(`
		<h1>%s</h1>
	`, txt)
}

func startSlide(index int) string {
	return fmt.Sprintf(`
		<div class="slide slide-%d">
		<div class="slide-content">
	`, index)
}
func endSlide() string {
	return `
		</div>
		</div>
	`
}

func getHighlightedMarkup(code string, lang string) (string, string) {

	lexer := lexers.Get(lang)

	if lexer == nil {
		log.Println("could not find correct lexer for", code)
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("swapoff")
	if style == nil {
		log.Println("using fallback styles")
		style = styles.Fallback
	}

	// formatter := formatters.Get("html")
	// formatter := formatters.Get("noop")
	formatter := html.New(html.WithClasses())
	// if formatter == nil {
	// 	formatter = formatters.Fallback
	// }

	highlightedCode := new(bytes.Buffer)
	cssClasses := new(bytes.Buffer)

	iterator, err := lexer.Tokenise(nil, code)
	err = formatter.Format(highlightedCode, style, iterator)
	if err != nil {
		log.Println("err", err)
		return fmt.Sprintf(`
			<pre>%s</pre>
		`, code), ""
	}

	formatter.WriteCSS(cssClasses, style)

	return highlightedCode.String(), cssClasses.String()
}
