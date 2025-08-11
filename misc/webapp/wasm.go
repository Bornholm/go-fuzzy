//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"slices"
	"sort"
	"syscall/js"

	"github.com/bornholm/go-fuzzy"
	"github.com/bornholm/go-fuzzy/dsl"
	"github.com/pkg/errors"
)

func main() {
	document := js.Global().Get("document")

	engine := fuzzy.NewEngine(fuzzy.Centroid(100))

	definitionTextArea := document.Call("getElementById", "definition")
	inputsTextArea := document.Call("getElementById", "inputs")
	executeButton := document.Call("getElementById", "execute")
	resultTextArea := document.Call("getElementById", "results")
	shareButton := document.Call("getElementById", "share")

	resetConsole := func() {
		resultTextArea.Set("value", "")
	}

	printConsole := func(text string) {
		resultTextArea.Set("value", resultTextArea.Get("value").String()+text+"\n")
		resultTextArea.Set("scrollTop", resultTextArea.Get("scrollHeight").Int())
	}

	logf := func(message string, args ...any) {
		printConsole(fmt.Sprintf(message, args...))
	}

	updateRulesAndVariables := func() bool {
		script := definitionTextArea.Get("value").String()
		if script == "" {
			return true
		}

		printConsole("Parsing definition...")

		result, err := dsl.ParseRulesAndVariables(script)
		if err != nil {
			printConsole(fmt.Sprintf("%+v", errors.WithStack(err)))
			return false
		}

		printConsole("Definition OK !")

		engine.Rules(result.Rules...)
		engine.Variables(result.Variables...)

		return true
	}

	var inputs fuzzy.Values

	updateInputs := func() bool {
		rawInputs := inputsTextArea.Get("value").String()
		if rawInputs == "" {
			return true
		}

		printConsole("Parsing inputs...")

		if err := json.Unmarshal([]byte(rawInputs), &inputs); err != nil {
			printConsole(fmt.Sprintf("%+v", errors.WithStack(err)))
			return false
		}

		printConsole("Inputs OK !")

		dumpValues(inputs, logf)

		return true
	}

	executeEngine := func() {
		resetConsole()

		if !updateRulesAndVariables() {
			return
		}
		if !updateInputs() {
			return
		}

		printConsole("Infering results...")

		results, err := engine.Infer(inputs)
		if err != nil {
			printConsole(fmt.Sprintf("%+v", errors.WithStack(err)))
			return
		}

		for _, v := range results.Variables() {
			dumpResult(engine, results, v, logf)
		}
	}

	executeButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		executeEngine()
		return nil
	}))

	shareButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		definition := definitionTextArea.Get("value").String()
		inputs := inputsTextArea.Get("value").String()

		payload := map[string]string{
			"d": definition,
			"i": inputs,
		}

		var buf bytes.Buffer

		encoder := gob.NewEncoder(&buf)

		if err := encoder.Encode(payload); err != nil {
			printConsole(fmt.Sprintf("%+v", errors.WithStack(err)))
			return nil
		}

		encodedPayload := base64.StdEncoding.EncodeToString(buf.Bytes())

		location := js.Global().Get("location")

		currentURL := location.Get("href").String()

		url, err := url.Parse(currentURL)
		if err != nil {
			printConsole(fmt.Sprintf("%+v", errors.WithStack(err)))
			return nil
		}

		query := url.Query()
		query.Set("p", encodedPayload)
		url.RawQuery = query.Encode()

		js.Global().Call("open", url.String(), "_blank")

		return nil
	}))

	location := js.Global().Get("location")
	currentURL := location.Get("href").String()

	url, err := url.Parse(currentURL)
	if err != nil {
		log.Printf("%+v", errors.WithStack(err))
	}

	if url != nil {
		if rawPayload := url.Query().Get("p"); rawPayload != "" {
			decoded, err := base64.StdEncoding.DecodeString(rawPayload)
			if err != nil {
				log.Printf("%+v", errors.WithStack(err))
			}
			if decoded != nil {
				buf := bytes.NewBuffer(decoded)
				decoder := gob.NewDecoder(buf)

				var payload map[string]string
				if err := decoder.Decode(&payload); err != nil {
					log.Printf("%+v", errors.WithStack(err))
				}

				if payload != nil {
					definitionTextArea.Set("value", payload["d"])
					definitionTextArea.Call("dispatchEvent", js.Global().Get("Event").New("change"))
					inputsTextArea.Set("value", payload["i"])
					inputsTextArea.Call("dispatchEvent", js.Global().Get("Event").New("change"))
				}
			}
		}
	}

	select {}
}

func dumpValues(values fuzzy.Values, logf func(message string, args ...any)) {
	logf("Values:")

	keys := slices.Collect(func(yield func(v string) bool) {
		for k := range values {
			if !yield(k) {
				return
			}
		}
	})

	sort.Strings(keys)

	for _, k := range keys {
		logf("|-> %s: %v", k, values[k])
	}
}

func dumpResult(engine *fuzzy.Engine, results fuzzy.Results, variable string, logf func(message string, args ...any)) {
	logf("Result: %s", variable)

	value, err := engine.Defuzzify(variable, results)
	if err != nil {
		panic(errors.WithStack(err))
	}

	logf("|")
	logf("|-> Value: %f", value)

	variableResults := results[variable]

	keys := slices.Collect(func(yield func(v string) bool) {
		for k := range variableResults {
			if !yield(k) {
				return
			}
		}
	})

	sort.Strings(keys)

	best := results.Best(variable)

	for _, term := range keys {
		isBest := ""
		if best.Term() == term {
			isBest = "(best)"
		}
		res := variableResults[term]
		logf("|--> %s %s", term, isBest)
		logf("|    |")
		logf("|    |--> TruthDegree: %f", res.TruthDegree())
		logf("|")
	}
}
