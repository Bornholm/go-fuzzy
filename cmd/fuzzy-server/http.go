package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/bornholm/go-fuzzy"
	"github.com/pkg/errors"
)

type jsonVariable struct {
	Name  string     `json:"name"`
	Terms []jsonTerm `json:"terms"`
}

type jsonTerm struct {
	Name       string    `json:"name"`
	Domain     []float64 `json:"domain"`
	Membership string    `json:"membership"`
}

// createHandler creates an HTTP handler for a specific fuzzy engine
func createHandler(registry *Registry) http.Handler {
	mux := http.NewServeMux()

	// Root endpoint - list available engines
	mux.HandleFunc("GET /api/v1/engines", func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Engines []string `json:"engines"`
		}{
			Engines: registry.Names(),
		}

		jsonResponse(w, response)
	})

	mux.HandleFunc("GET /api/v1/engines/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		// Check if engine exists
		variables, _, exists := registry.Get(name)
		if !exists {
			http.Error(w, fmt.Sprintf("Engine '%s' not found", name), http.StatusNotFound)
			return
		}

		response := struct {
			Variables []jsonVariable `json:"variables"`
			Rules     []string       `json:"rules"`
		}{
			Variables: make([]jsonVariable, 0),
			Rules:     make([]string, 0),
		}

		for _, v := range variables {
			terms := slices.Collect[jsonTerm](func(yield func(jsonTerm) bool) {
				for _, t := range v.Terms() {
					min, max := t.Domain()
					term := jsonTerm{
						Name:       t.Name(),
						Domain:     []float64{min, max},
						Membership: fmt.Sprintf("%T", t.Membership()),
					}
					if !yield(term) {
						return
					}
				}
			})
			slices.SortFunc(terms, func(a, b jsonTerm) int {
				return strings.Compare(a.Name, b.Name)
			})
			response.Variables = append(response.Variables, jsonVariable{
				Name:  v.Name(),
				Terms: terms,
			})
		}

		jsonResponse(w, response)
	})

	mux.HandleFunc("POST /api/v1/engines/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		// Check if engine exists
		variables, rules, exists := registry.Get(name)
		if !exists {
			http.Error(w, fmt.Sprintf("Engine '%s' not found", name), http.StatusNotFound)
			return
		}

		defuzz := r.URL.Query().Get("defuzz")
		if defuzz == "" {
			defuzz = "centroid"
		}

		rawSteps := r.URL.Query().Get("steps")
		if rawSteps == "" {
			rawSteps = "100"
		}

		steps, err := strconv.ParseInt(rawSteps, 10, 32)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid step value '%v', expected integer", steps), http.StatusBadRequest)
			return
		}

		var defuzzify fuzzy.DefuzzifyFunc

		switch defuzz {
		case "centroid":
			defuzzify = fuzzy.Centroid(int(steps))
		case "mean-max":
			defuzzify = fuzzy.MeanOfMaximum(int(steps))
		default:
			http.Error(w, fmt.Sprintf("Invalid defuzzification function '%s'", name), http.StatusBadRequest)
			return
		}

		engine := fuzzy.NewEngine(defuzzify)
		engine.Variables(variables...)
		engine.Rules(rules...)

		// Parse JSON input
		var inputValues fuzzy.Values
		if err := json.NewDecoder(r.Body).Decode(&inputValues); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		// Run inference
		results, err := engine.Infer(inputValues)
		if err != nil {
			http.Error(w, fmt.Sprintf("Inference error: %v", err), http.StatusInternalServerError)
			return
		}

		type jsonTermResult struct {
			TruthDegree float64 `json:"truthDegree"`
		}

		type jsonVariableResult struct {
			Value float64                   `json:"value"`
			Best  string                    `json:"best,omitempty"`
			Terms map[string]jsonTermResult `json:"terms,omitempty"`
		}

		// Prepare response
		response := struct {
			Results map[string]jsonVariableResult `json:"results"`
		}{
			Results: make(map[string]jsonVariableResult),
		}

		// Process results for each variable
		for varName, varResults := range results {
			jsonVar := jsonVariableResult{
				Terms: make(map[string]jsonTermResult),
			}

			// Find the best term
			bestTerm, ok := results.Best(varName)
			if ok {
				jsonVar.Best = bestTerm.Term()
			}

			// Get defuzzified value if possible
			if len(varResults) > 0 {
				defuzz, err := engine.Defuzzify(varName, results)
				if err != nil {
					http.Error(w, fmt.Sprintf("Could not defuzzify value: %+v", errors.WithStack(err)), http.StatusInternalServerError)
					return
				}

				jsonVar.Value = defuzz
			}

			// Add results for each term
			for termName, result := range varResults {
				termResult := jsonTermResult{
					TruthDegree: result.TruthDegree(),
				}

				jsonVar.Terms[termName] = termResult
			}

			response.Results[varName] = jsonVar
		}

		jsonResponse(w, response)
	})

	return mux
}

func jsonResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(response); err != nil {
		log.Printf("[ERROR] could not encode response: %+v", errors.WithStack(err))
	}
}

// LoggingMiddleware logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
