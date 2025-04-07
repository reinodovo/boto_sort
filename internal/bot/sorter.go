package bot

func mergeSort(items []string, c *Comparator, chatId int64, result chan SortingResult) {
	sortedItems := mergeSortRecursive(items, c, chatId)
	result <- SortingResult{
		chatId:      chatId,
		sortedItems: sortedItems,
	}
}

func mergeSortRecursive(items []string, c *Comparator, chatId int64) []string {
	if len(items) <= 1 {
		return items
	}
	mid := len(items) / 2
	left := mergeSortRecursive(items[:mid], c, chatId)
	right := mergeSortRecursive(items[mid:], c, chatId)
	return merge(left, right, c, chatId)
}

func merge(left, right []string, c *Comparator, chatId int64) []string {
	result := make([]string, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if c.Compare(left[i], right[j], chatId) < 0 {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}
