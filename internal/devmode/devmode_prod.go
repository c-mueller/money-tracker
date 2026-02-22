//go:build !dev

package devmode

const Enabled = false

func SetupUser(_ func() (int, error)) (int, error) {
	return 0, nil
}
