package input

import (
	"regexp"
	"unicode"

	"github.com/veandco/go-sdl2/sdl"
)

var pressedKeys map[rune]bool

// holds keys like ctrl, shift, enter, ...
var pressedSpecialKeys map[int]bool

var lastPressedKey rune

var graphicalCharRegex *regexp.Regexp

func init() {
	graphicalCharRegex = regexp.MustCompile(`[[:graph:]]| `)
	pressedKeys = make(map[rune]bool)
	pressedSpecialKeys = make(map[int]bool)
}

// PressKey notifies the input package that the key was pressed
func PressKey(keycode sdl.Keycode) (isevent bool) {
	char := rune(int(keycode))

	// if the char has a graphical representation
	if graphicalCharRegex.MatchString(string(char)) {
		pressedKeys[char] = true
		lastPressedKey = char
		return true
	}

	pressedSpecialKeys[int(keycode)] = true
	return false
}

// ReleaseKey notifies the input package that the key was released
func ReleaseKey(keycode sdl.Keycode) {
	char := rune(int(keycode))

	if graphicalCharRegex.MatchString(string(char)) {
		pressedKeys[char] = false
		lastPressedKey = 0
	} else {
		pressedSpecialKeys[int(keycode)] = false
	}
}

// GetPressedKeys gives back a list of currently pressed keys
func GetPressedKeys() (keys []rune) {
	for r, b := range pressedKeys {
		if b {
			keys = append(keys, r)
		}
	}

	return
}

// IsSpecialKeyPressed tests if a special key (shift, ctrl, ...) is pressed
func IsSpecialKeyPressed(keycode sdl.Keycode) bool {
	if _, ok := pressedSpecialKeys[int(keycode)]; ok {
		return pressedSpecialKeys[int(keycode)]
	}

	return false
}

// GetLastPressedKey returns the last pressed non-special key
func GetLastPressedKey() (key rune) {
	key = lastPressedKey

	// capitalize if shift is pressed
	if IsSpecialKeyPressed(sdl.K_LSHIFT) ||
		IsSpecialKeyPressed(sdl.K_RSHIFT) {
		key = unicode.ToUpper(key)
	}

	return
}
