package main

type RipgrepResult = map[string][]RipgrepMatch

type RipgrepMatch struct {
	Path        string
	Row         int
	Col         int
	MatchedLine string
}

func (a *App) Search(search_term string, dir string, include string, exclude string) RipgrepResult {
	if search_term == "" {
		return RipgrepResult{}
	}
	matches := Ripgrep(search_term, dir, include, exclude)
	grouped := GroupByProperty(matches, func(match RipgrepMatch) string {
		return match.Path
	})
	return grouped
}

func GroupByProperty[T any, K comparable](items []T, getProperty func(T) K) map[K][]T {
	grouped := make(map[K][]T)

	for _, item := range items {
		key := getProperty(item)
		grouped[key] = append(grouped[key], item)
	}

	return grouped
}
