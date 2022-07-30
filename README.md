
![test status](https://github.com/fatanugraha/noloopclosure/actions/workflows/test.yaml/badge.svg)

# noloopclosure

`noloopclosure` is a linter that disallow reference capture of loop variable inside of a closure.

This linter can prevent bugs introduced by [a very popular gotcha in Go][1] by disallowing implicit reference capture when creating closure inside a loop. 

## Installation
`go install github.com/fatanugraha/noloopclosure/cmd/noloopclosure@latest`

## Difference with `go vet`'s `loopclosure`
`loopclosure` will only complain if the captured loop variable is inside of a function closure that preceded by `go` and `defer` keyword and is the last statement in the loop's block ([reference][2]).

This linter complain each time it found any captured loop variable inside a function closure. This linter is helpful if you have utilities that abstracts the `go` behavior, for example:
```
func main() {
	for i := 0; i < 10; i++ {
		runInGoroutine(func() { fmt.Println(i) })
	}
}

func runInGoroutine(f func()) {
	go f()
}
```

will pass the `go vet`'s check while fails the `noloopclosure` check.

It's generally a good idea (unless every bit of performance matters) to extract the function creation part inside your 
loop so that you don't accidentally capture the reference and cause unwanted bugs. 
For example, code above can be re-written as:

```
func main() {
	for i := 0; i < 10; i++ {
		runInGoroutine(newIntPrinter(i))
	}
}

func newIntPrinter(i int) func() {
    func() { fmt.Println(i) }
}

func runInGoroutine(f func()) {
	go f()
}
```

[1]: https://go.dev/doc/faq#closures_and_goroutines
[2]: https://cs.opensource.google/go/x/tools/+/refs/tags/v0.1.12:go/analysis/passes/loopclosure/loopclosure.go;l=19
