
![test status](https://github.com/fatanugraha/noloopclosure/actions/workflows/test.yaml/badge.svg)

# noloopclosure

`noloopclosure` is a linter that disallow reference capture of loop variable inside of a closure.

This linter can prevent bugs introduced by [a very popular gotcha in Go][1] by disallowing implicit reference capture when creating closure inside a loop. 

## Installation
`https://github.com/fatanugraha/noloopclosure`

## Difference with `go vet`'s `loopclosure`
`loopclosure` will only complain if the captured loop variable is inside of a function closure that preceded by `go` and `defer` keyword and is the last statement in the loop's block ([reference][2]).

This linter complain each time it found any captured loop variable inside a function closure. This linter is helpful if you have utilities that abstracts the `go` behavior, for example:
```
func runInGoroutine(f func()) {
	go f()
}

func main() {
	for i := 0; i < 10; i++ {
		runInGoroutine(func() { fmt.Println(i) })
	}
}
```

will pass the `go vet`'s check while fails the `noloopclosure` check.


[1]: https://go.dev/doc/faq#closures_and_goroutines
[2]: https://cs.opensource.google/go/x/tools/+/refs/tags/v0.1.12:go/analysis/passes/loopclosure/loopclosure.go;l=19
