package util

type SortableString []string

func (s SortableString) Len() int {
	return len(s)
}

func (s SortableString) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s SortableString) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
