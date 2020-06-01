package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
)

// readLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// readLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

func PathExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// stringsHas checks the target string slice contains src or not
func StringsHas(target []string, src string) bool {
	for _, t := range target {
		if strings.TrimSpace(t) == src {
			return true
		}
	}
	return false
}

// formatBytes format the
func FormatBytes(bytes uint64, decimals uint8) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	const k float64 = 1000
	var sizes = [...]string{"Bytes", "KB", "MB", "GB", "TB", "PB"}

	var i float64 = math.Floor(math.Log(float64(bytes)) / math.Log(float64(k)))

	return fmt.Sprintf(fmt.Sprintf("%%.%vf %%s", i), float64(bytes)/math.Pow(k, i), sizes[int(i)])
}

// Respond with Json data
func Respond(w http.ResponseWriter, status int, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
