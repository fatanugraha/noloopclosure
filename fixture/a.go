package fixture

func lol() {
	for _, i := range []int{1, 2, 3, 4} {
		g := func() int {
			return i
		}

		g()
	}

	i := 0
	for i++; i < 5; i++ {
	}
}
