package utils

func RemoveOwnOccupiedSquares(a, b []string) []string {
    // build a set from b
    removeSet := make(map[string]struct{}, len(b))
    for _, x := range b {
        removeSet[x] = struct{}{}
    }

    result := make([]string, 0, len(a))
    for _, x := range a {
        if _, found := removeSet[x]; !found {
            result = append(result, x)
        }
    }
    return result
}