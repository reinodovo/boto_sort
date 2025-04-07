package bot

func heapify(arr []string, n, i int, c *Comparator, chatId int64) {
	largest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && c.Compare(arr[largest], arr[left], chatId) < 0 {
		largest = left
	}
	if right < n && c.Compare(arr[largest], arr[right], chatId) < 0 {
		largest = right
	}

	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, n, largest, c, chatId)
	}
}

func sort(arr []string, c *Comparator, chatId int64, sortedItem chan SortedItem, finishedSorting chan FinishedSorting) {
	n := len(arr)

	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i, c, chatId)
	}

	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		sortedItem <- SortedItem{
			chatId:   chatId,
			position: i + 1,
			item:     arr[i],
		}
		heapify(arr, i, 0, c, chatId)
	}

	sortedItem <- SortedItem{
		chatId:   chatId,
		position: 0,
		item:     arr[0],
	}
	finishedSorting <- FinishedSorting{
		chatId: chatId,
		items:  arr,
	}
}
