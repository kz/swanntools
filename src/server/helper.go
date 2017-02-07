package main

func intInSlice(x *int, list *[]int) bool {
	for _, item := range *list {
		if item == *x {
			return true
		}
	}
	return false
}

func intPositionInSlice(x *int, list *[]int) (int, bool) {
	for i, item := range *list {
		if item == *x {
			return  i, true
		}
	}
	return -1, false
}
