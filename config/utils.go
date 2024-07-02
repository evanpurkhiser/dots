package config

// listDifference computes the list of items that exist in list1 without items
// that exist in list2.
func listDifference(list1, list2 []string) []string {
	list := []string{}

	for _, val := range list1 {
		intersected := false

		for _, val2 := range list2 {
			if val2 == val {
				intersected = true
				break
			}
		}

		if !intersected {
			list = append(list, val)
		}
	}

	return list
}

// listIntersect computes the list of items that exist in both list{1,2}.
func listIntersect(list1, list2 []string) []string {
	list := []string{}

	for _, val := range list1 {
		for _, val2 := range list2 {
			if val2 == val {
				list = append(list, val)
				break
			}
		}
	}

	return list
}

// removeDupes removes duplicate entries from a list, returning the list
// without duplicates, and a list of found duplicates.
func removeDupes(list []string) ([]string, []string) {
	newList := []string{}
	removed := []string{}

	dupesMap := map[string]bool{}

	// 2. Check for duplicate groups in the list
	for _, group := range list {
		if !dupesMap[group] {
			dupesMap[group] = true

			newList = append(newList, group)
			continue
		}

		// Remove duplicate group from the list
		removed = append(removed, group)
	}

	return newList, removed
}
