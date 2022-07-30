package fixtures

import "fmt"

func ForLoop() {
	for i := 0; i < 5; i++ {
		_ = i
	}

	for false {
	}

	for i := 0; i < 5; i++ {
		_ = func() {
			fmt.Println(i) // want "found reference to loop variable `i`. Consider to duplicate variable `i` before using it inside the function closure."
		}
	}

	k := 5
	for i, j := 0, 0; i < j; i++ {
		_ = func() {
			fmt.Println(k)
		}

		_ = func() {
			fmt.Println(i, j) // want "found reference to loop variable `i`. Consider to duplicate variable `i` before using it inside the function closure." "found reference to loop variable `j`. Consider to duplicate variable `j` before using it inside the function closure."
		}
	}
}

func RangeLoop() {
	for k, v := range map[string]int{} {
		_ = func() {
			fmt.Println(k, v) // want "found reference to loop variable `k`. Consider to duplicate variable `k` before using it inside the function closure." "found reference to loop variable `v`. Consider to duplicate variable `v` before using it inside the function closure."
		}
	}

	for _, v := range map[string]int{} {
		_ = func() {
			fmt.Println(v) // want "found reference to loop variable `v`. Consider to duplicate variable `v` before using it inside the function closure."
		}
	}

	for k := range map[string]int{} {
		_ = func() {
			fmt.Println(k) // want "found reference to loop variable `k`. Consider to duplicate variable `k` before using it inside the function closure."
		}
	}

	for range map[string]int{} {
	}
}
