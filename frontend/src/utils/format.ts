export const formatSize = (size: number) => {
    if (!size || size <= 0) return '0 B'

    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let value = size
    let index = 0

    while (value >= 1024 && index < units.length - 1) {
        value /= 1024
        index++
    }

    return `${value.toFixed(2)} ${units[index]}`
}

export const formatToken = (tokens: number) => {
    if (!tokens || tokens <= 0) return '0'

    if (tokens >= 10000) {
        return `${(tokens / 1000).toFixed(1)}k`
    }

    if (tokens >= 1000) {
        return `${(tokens / 1000).toFixed(2)}k`
    }

    return String(tokens)
}

export const getTokenLevel = (tokens: number) => {
    if (tokens >= 20000) return 'danger'
    if (tokens >= 8000) return 'warning'
    return 'info'
}