import type { TreeNode } from '../types/gingest'

export const escapeXmlAttribute = (value: string) => {
    return value
        .replaceAll('&', '&amp;')
        .replaceAll('"', '&quot;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
}

export const safeCDATA = (value: string) => {
    return value.replaceAll(']]>', ']]]]><![CDATA[>')
}

export const generateTreeText = (nodes: TreeNode[], prefix = ''): string => {
    let text = ''

    nodes.forEach((node, index) => {
        const isLast = index === nodes.length - 1
        const connector = isLast ? '└── ' : '├── '
        const childPrefix = prefix + (isLast ? '    ' : '│   ')

        text += prefix + connector + node.label + (node.isFile ? '' : '/') + '\n'

        if (node.children && node.children.length > 0) {
            text += generateTreeText(node.children, childPrefix)
        }
    })

    return text
}

export const buildXmlByFiles = (
    projectName: string,
    directoryTree: TreeNode[],
    files: TreeNode[],
    exportType: string,
) => {
    const estimatedTokens = Math.floor(
        files.reduce((sum, file) => sum + (file.content?.length || 0), 0) / 4,
    )

    let xml = ''
    xml += '<project_summary>\n'
    xml += `Project: ${projectName}\n`
    xml += `Export Type: ${exportType}\n`
    xml += `Total Files: ${files.length}\n`
    xml += `Estimated Tokens: ${estimatedTokens}\n`
    xml += '</project_summary>\n\n'

    xml += '<directory_tree>\n'
    xml += '.\n'
    xml += generateTreeText(directoryTree)
    xml += '</directory_tree>\n\n'

    xml += '<files>\n'

    files.forEach((file) => {
        xml += `<file path="${escapeXmlAttribute(file.fullPath || file.label)}">\n`
        xml += '<![CDATA[\n'
        xml += safeCDATA(file.content || '')
        xml += '\n]]>\n'
        xml += '</file>\n\n'
    })

    xml += '</files>'

    return xml
}

export const buildSuggestedXmlFileName = (projectName: string, selected: boolean) => {
    if (!projectName) return 'gingest_export.xml'

    const suffix = selected ? '_selected' : '_full'

    return (
        projectName
            .replace(/^Local:\s*/, '')
            .replace(/[\\/:*?"<>|]/g, '_') +
        `_gingest${suffix}.xml`
    )
}