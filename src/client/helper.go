package main

// intInSlice checks if an integer is in a slice
func intInSlice(x *int, list *[]int) bool {
	for _, item := range *list {
		if item == *x {
			return true
		}
	}
	return false
}
