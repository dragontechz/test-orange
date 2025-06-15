package utils

import (
	"fmt"
	"strings"
)

func main() {
	data := "HTTP/1.1 200 Bytes1xxxBytes0\r\n\r\n"
	start := "Bytes1"
	end := "Bytes0"

	fmt.Printf(GET(data, start, end))
	fmt.Printf(Insert(data, "fff", start, end))

}

func GET(data, start, end string) (string, error) {
	startIndex := strings.Index(data, start)
	if startIndex == -1 {
		return "", fmt.Errorf("délimiteur de début '%s' non trouvé", start)
	}
	startIndex += len(start) // On se place après le délimiteur

	endIndex := strings.Index(data[startIndex:], end)
	if endIndex == -1 {
		return "", fmt.Errorf("délimiteur de fin '%s' non trouvé", end)
	}
	endIndex += startIndex // Ajuste l'indice par rapport à la chaîne complète

	// Vérifie que les indices sont valides
	if startIndex > endIndex || endIndex > len(data) {
		return "", fmt.Errorf("indices invalides (start=%d, end=%d)", startIndex, endIndex)
	}

	return data[startIndex:endIndex], nil
}
func Insert(data, input, start, end string) string {
	start_index := strings.Index(data, start) + len(start)
	end_index := strings.Index(data, end)
	resp := data[:start_index] + input + data[end_index:]
	return resp
}
