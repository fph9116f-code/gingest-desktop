package ingest

import (
	"sort"
	"strings"

	"gingest-desktop/internal/model"
)

type helperNode struct {
	Label    string
	IsFile   bool
	FullPath string
	Content  string
	ChildMap map[string]*helperNode
}

func BuildDirectoryTree(paths []string, contentMap map[string]string) []model.TreeNode {
	root := &helperNode{
		Label:    "root",
		ChildMap: map[string]*helperNode{},
	}

	for _, path := range paths {
		parts := strings.Split(path, "/")
		current := root

		for i, part := range parts {
			if part == "" {
				continue
			}

			if current.ChildMap[part] == nil {
				current.ChildMap[part] = &helperNode{
					Label:    part,
					ChildMap: map[string]*helperNode{},
				}
			}

			current = current.ChildMap[part]

			if i == len(parts)-1 {
				current.IsFile = true
				current.FullPath = path
				current.Content = contentMap[path]
			}
		}
	}

	compressTree(root)

	result := childrenToList(root)
	assignIDs(result, 1)

	return result
}

func compressTree(node *helperNode) {
	for _, child := range node.ChildMap {
		compressTree(child)
	}

	if node.Label != "root" && !node.IsFile && len(node.ChildMap) == 1 {
		var singleChild *helperNode
		for _, child := range node.ChildMap {
			singleChild = child
			break
		}

		node.Label = node.Label + "/" + singleChild.Label
		node.IsFile = singleChild.IsFile
		node.FullPath = singleChild.FullPath
		node.Content = singleChild.Content
		node.ChildMap = singleChild.ChildMap
	}
}

func childrenToList(node *helperNode) []model.TreeNode {
	keys := make([]string, 0, len(node.ChildMap))
	for key := range node.ChildMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]model.TreeNode, 0, len(keys))

	for _, key := range keys {
		child := node.ChildMap[key]

		treeNode := model.TreeNode{
			Label:    child.Label,
			IsFile:   child.IsFile,
			FullPath: child.FullPath,
			Content:  child.Content,
			Children: childrenToList(child),
		}

		if len(treeNode.Children) == 0 {
			treeNode.Children = nil
		}

		result = append(result, treeNode)
	}

	return result
}

func assignIDs(nodes []model.TreeNode, start int) int {
	current := start

	for i := range nodes {
		nodes[i].ID = current
		current++

		if len(nodes[i].Children) > 0 {
			current = assignIDs(nodes[i].Children, current)
		}
	}

	return current
}
