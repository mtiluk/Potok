package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var ErrEmptyInput = errors.New("input cannot be empty")

func Input(promptText string) (string, error) {
	return inputFrom(promptText, os.Stdin, true)
}

func InputDefault(label, def string) (string, error) {
	text, err := inputFrom(fmt.Sprintf("%s [%s]: ", label, def), os.Stdin, false)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" {
		return def, nil
	}
	return text, nil
}

func Secret(promptText string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return inputFrom(promptText, os.Stdin, true)
	}

	for {
		fmt.Print(promptText)
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		fmt.Println()

		s := strings.TrimSpace(string(b))
		if s == "" {
			fmt.Println("Input cannot be empty.")
			continue
		}
		return s, nil
	}
}

func inputFrom(promptText string, r io.Reader, requireNonEmpty bool) (string, error) {
	reader := bufio.NewReader(r)

	for {
		fmt.Print(promptText)

		input, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		input = strings.TrimSpace(input)

		if requireNonEmpty && input == "" {
			fmt.Println("Input cannot be empty.")
			continue
		}

		return input, nil
	}
}
