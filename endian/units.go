package endian

import "fmt"

func IEEE1541frombits(n uint64) string {
	const prefix = " KMGTPEZ"

	i := 0
	m := n
	for n > 1024 && i < len(prefix) {
		i++
		m = n
		n /= 1024
	}

	if prefix[i] == ' ' {
		return fmt.Sprintf("%d bytes", n)
	}

	units := fmt.Sprintf("%ciB", prefix[i])
	d := (float64(m) - float64(n*1024) + 51.2) / 102.4
	if d >= 10 {
		d = 0
		n++
	}
	return fmt.Sprintf("%d.%d %s", n, uint64(d), units)
}
