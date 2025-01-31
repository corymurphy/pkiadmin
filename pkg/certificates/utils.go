package certificates

import "fmt"

// TODO name this properly
func PrintBytes(bytes []byte) {
	fmt.Println("Dumping response data:")
	for i := 0; i < len(bytes); i += 16 {
		end := i + 16
		if end > len(bytes) {
			end = len(bytes)
		}
		fmt.Printf("%04x: ", i)
		for j := i; j < end; j++ {
			fmt.Printf("%02x ", bytes[j])
		}
		fmt.Println()
	}

	fmt.Println()
}
