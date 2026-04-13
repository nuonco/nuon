import { cn } from '@/utils/classnames'

export const markdownStyles = `
  .mermaid-diagram {
    text-align: center;
    margin: 1rem 0;
    min-height: 100px;
    border-radius: 0.25rem;
    border: 1px solid var(--border-color);
    padding: 1rem;
    background: var(--background-neutral);
  }
  .mermaid-diagram svg {
    max-width: 100%;
    height: auto;
  }
  details[open] > summary {
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 0;
  }
  details div ul, details div ol {
    padding-left: 1.5rem;
  }
  .prose :where(code):not(:where([class~="not-prose"], [class~="not-prose"] *))::before,
  .prose :where(code):not(:where([class~="not-prose"], [class~="not-prose"] *))::after,
  .prose code::before,
  .prose code::after {
    content: none !important;
  }
  .prose andypf-json-viewer *[class*="container"],
  .prose andypf-json-viewer *[class="container"],
  .prose andypf-json-viewer div.container,
  .prose andypf-json-viewer div.container *,
  .prose div andypf-json-viewer,
  .prose div andypf-json-viewer .container,
  .prose div andypf-json-viewer .container * {
    font-family: var(--font-hack) !important;
    font-size: 0.875rem !important;
    line-height: 1.25rem !important;
  }
  andypf-json-viewer .container,
  andypf-json-viewer .container * {
    font-family: var(--font-hack) !important;
    font-size: 0.875rem !important;
    line-height: 1.25rem !important;
  }
  .prose andypf-json-viewer {
    --font-family: var(--font-hack) !important;
    --font-size: 0.875rem !important;
    --line-height: 1.25rem !important;
    --json-font-family: var(--font-hack) !important;
    --json-font-size: 0.875rem !important;
    --json-line-height: 1.25rem !important;
    --viewer-font-family: var(--font-hack) !important;
    --viewer-font-size: 0.875rem !important;
    --text-font-size: 0.875rem !important;
    --code-font-family: var(--font-hack) !important;
    --code-font-size: 0.875rem !important;
    --container-font-family: var(--font-hack) !important;
    --container-font-size: 0.875rem !important;
    --key-value-font-family: var(--font-hack) !important;
    --key-value-font-size: 0.875rem !important;
    --value-string-font-family: var(--font-hack) !important;
    --value-string-font-size: 0.875rem !important;
  }
`

export const proseClassName = cn(
  'prose dark:prose-invert max-w-[100%]',
  'prose-code:bg-code prose-code:text-sm prose-code:text-blue-500 prose-code:font-mono'
)
