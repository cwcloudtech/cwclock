# AI instruction 2

## UX/UI

Improve the tooltip UI design (for now it seems it's just titles and are not shown when hovering on the sidebar).

The color picker must display the selected color surligning the hexadecimal value.

## Frontend refactoring

Cancel secondary buttons must be inside the `ConfigForm` webcomponent and display in the same line on the right of the submit button.

## Backend refactoring

I want every test `!= ""` or `== ""` to be replaced by this utils functions in a utils package:

```go
func IsNotBlank(str string) bool {
	return len(str) > 0 && strings.TrimSpace(str) != ""
}

func IsBlank(str string) bool {
	return !IsNotBlank(str)
}
```

And also test with default values to use this (same utils package):

```go
func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}
```
