//go:build dev

package devmode

const Enabled = true

func SetupUser(getOrCreate func() (int, error)) (int, error) {
	return getOrCreate()
}
