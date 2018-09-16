package input

import "github.com/veandco/go-sdl2/sdl"

type Key struct{}

var pressedKeys map[string]Key = make(map[string]Key)

func PressKey(keycode sdl.Keycode) {
	char := string(int(keycode))

	pressedKeys[char] = Key{}
}

func ReleaseKey(keycode sdl.Keycode) {
	char := string(int(keycode))

	delete(pressedKeys, char)
}

func GetPressedKeys() (keys []string) {
	for _, k := range keys {
		keys = append(keys, k)
	}

	return
}
