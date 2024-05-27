package frontend

import "strings"

type MenuRoute struct {
	Title        string
	Path         string
	Icon         string
	AllowedRoles []string
	Items        []MenuRoute
}

func IsActiveItemElement(itemPath, currentPath string, isBase bool) bool {
	if isBase {
		return itemPath == currentPath
	}

	return strings.HasPrefix(currentPath, itemPath)
}
