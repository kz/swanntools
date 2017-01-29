package main

func intInSlice(x *int, list *[]int) bool {
	for _, item := range *list {
		if item == *x {
			return true
		}
	}
	return false
}
