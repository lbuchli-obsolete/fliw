package backend

import (
	"testing"
)

func TestInit(t *testing.T) {
	Init(nil)
}

var calculateValueTest = map[string]string{
	`(6/2)*((9+11)-(3*3)+5)`: `48`,
	`3*3`:                    `9`,
	`3.1*(4/3)*5^3`:          `517`,
	`81/3%`:                  `27%`,
}

func TestCalculateValue(t *testing.T) {
	for op, expected := range calculateValueTest {
		result := CalculateValue(op, nil)

		if result != expected {
			t.Error("Expected ", expected, ", gave ", result)
		}
	}
}
