import { useState } from 'react'

interface IJsonViewer {
  data: any
  collapsed?: boolean
}

export const JsonViewer = ({ data, collapsed = false }: IJsonViewer) => {
  const [isCollapsed, setIsCollapsed] = useState(collapsed)

  const formatted = typeof data === 'string' ? data : JSON.stringify(data, null, 2)

  if (isCollapsed) {
    return (
      <button
        onClick={() => setIsCollapsed(false)}
        className="text-xs text-primary-600 hover:text-primary-800 dark:text-primary-400 dark:hover:text-primary-300"
      >
        Show JSON ({typeof data === 'object' ? Object.keys(data || {}).length : 0} keys)
      </button>
    )
  }

  return (
    <div className="relative">
      {collapsed && (
        <button
          onClick={() => setIsCollapsed(true)}
          className="absolute right-2 top-2 text-xs text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          Collapse
        </button>
      )}
      <pre className="overflow-x-auto rounded-md bg-gray-900 p-4 text-xs text-gray-100 dark:bg-black dark:text-gray-200">
        <code>{formatted}</code>
      </pre>
    </div>
  )
}
