package util

func StringDiff(new, old []string) (added []string, removed []string) {
	encountered := map[string]int{}
	added = []string{}
	removed = []string{}

	for _, tag := range old {
		if _, ok := encountered[tag]; !ok {
			encountered[tag]++
		}
	}
	for _, tag := range new {
		encountered[tag]--
	}
	for name, flag := range encountered {
		if flag == -1 {
			added = append(added, name)
		}
		if flag == 1 {
			removed = append(removed, name)
		}
	}
	return added, removed
}
