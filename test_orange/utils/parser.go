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

func GET(data, start, end string) string {
	start_index := strings.Index(data, start) + len(start)
	end_index := strings.Index(data, end)
	resp := data[start_index:end_index]
	return resp
}

func Insert(data, input, start, end string) string {
	start_index := strings.Index(data, start) + len(start)
	end_index := strings.Index(data, end)
	resp := data[:start_index] + input + data[end_index:]
	return resp
}
