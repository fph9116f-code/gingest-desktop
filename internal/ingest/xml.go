package ingest

import (
	"html"
	"strings"

	"gingest-desktop/internal/model"
)

func BuildFullXML(response model.GingestResponse) string {
	var sb strings.Builder

	sb.WriteString("<project_summary>\n")
	sb.WriteString("Project: ")
	sb.WriteString(response.ProjectName)
	sb.WriteString("\n")
	sb.WriteString("Total Files: ")
	sb.WriteString(intToString(response.FileCount))
	sb.WriteString("\n")
	sb.WriteString("Estimated Tokens: ")
	sb.WriteString(int64ToString(response.EstimatedTokens))
	sb.WriteString("\n")
	sb.WriteString("</project_summary>\n\n")

	sb.WriteString("<directory_tree>\n")
	sb.WriteString(".\n")
	sb.WriteString(GenerateTreeText(response.DirectoryTree, ""))
	sb.WriteString("</directory_tree>\n\n")

	sb.WriteString("<files>\n")
	appendFiles(&sb, response.DirectoryTree)
	sb.WriteString("</files>")

	return sb.String()
}

func GenerateTreeText(nodes []model.TreeNode, prefix string) string {
	var sb strings.Builder

	for i, node := range nodes {
		isLast := i == len(nodes)-1
		connector := "├── "
		childPrefix := prefix + "│   "

		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		sb.WriteString(prefix)
		sb.WriteString(connector)
		sb.WriteString(node.Label)

		if !node.IsFile {
			sb.WriteString("/")
		}

		sb.WriteString("\n")

		if len(node.Children) > 0 {
			sb.WriteString(GenerateTreeText(node.Children, childPrefix))
		}
	}

	return sb.String()
}

func appendFiles(sb *strings.Builder, nodes []model.TreeNode) {
	for _, node := range nodes {
		if node.IsFile {
			sb.WriteString(`<file path="`)
			sb.WriteString(html.EscapeString(node.FullPath))
			sb.WriteString(`">`)
			sb.WriteString("\n")
			sb.WriteString("<![CDATA[\n")
			sb.WriteString(safeCDATA(node.Content))
			sb.WriteString("\n]]>\n")
			sb.WriteString("</file>\n\n")
			continue
		}

		if len(node.Children) > 0 {
			appendFiles(sb, node.Children)
		}
	}
}

func safeCDATA(content string) string {
	return strings.ReplaceAll(content, "]]>", "]]]]><![CDATA[>")
}

func intToString(value int) string {
	return int64ToString(int64(value))
}

func int64ToString(value int64) string {
	if value == 0 {
		return "0"
	}

	var digits []byte
	negative := value < 0
	if negative {
		value = -value
	}

	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
