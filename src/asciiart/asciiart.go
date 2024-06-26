package asciiart

import (
	"bufio"
	"os"
	"strings"
)

func GetAsciiLine(filename string, num int) (string, error) {
	file, e := os.Open(filename)
	if e != nil {
		return "", e
	}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	line := ""
	for scanner.Scan() {
		if lineNum == num {
			line = scanner.Text()
		}
		lineNum++
	}
	return line, nil
}

func AsciiArt(input, filename string) (string, error) {

	banner := "banners/" + filename + ".txt"
	line := ""
	result := "\n"

	args := strings.Split(input, "\r\n")

	for _, word := range args {
		for i := 0; i < 8; i++ {
			for _, letter := range word {
				asciiLine, err := GetAsciiLine(banner, 1+int(letter-' ')*9+i)
				if err != nil {
					return "", err
				}

				result += asciiLine
			}
			if word == "" && i == 0 {
				line += "\n"
			} else if word == "" {

			} else {
				line += "\n"
			}
			result += line
			line = ""
		}
	}

	return result, nil
}
