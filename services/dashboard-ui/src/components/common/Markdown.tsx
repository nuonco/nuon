import showdown from 'showdown'
import { cn } from '@/utils/classnames'

showdown.extension('targetlink', () => {
  return [
    {
      type: 'output',
      regex: /<a\s+href="(?!#)(.*?)"(.*?)>/g,
      replace: '<a href="$1" target="_blank" $2>',
    },
  ]
})

showdown.extension('wrapTables', () => [
  {
    type: 'output',
    filter: (text: string) =>
      text.replace(
        /(<table[^>]*>[\s\S]*?<\/table>)/g,
        '<div class="readme-table">$1</div>'
      ),
  },
])

const markdown = new showdown.Converter({
  extensions: ['targetlink', 'wrapTables'],
  tables: true,
  tasklists: true,
})

export const Markdown = ({ content = '' }) => (
  <>
    <style>{`.prose .readme-table pre { max-width: 50ch; }`}</style>
    <div
      className={cn(
        'prose dark:prose-invert max-w-[100%]',
        'prose-code:bg-code prose-code:text-sm prose-code:text-blue-500 prose-code:font-mono',
        'prose-pre:bg-code prose-pre:text-sm prose-pre:text-blue-500 prose-pre:font-mono prose-pre:rounded prose-pre:shadow-sm prose-pre:overflow-auto prose-pre:max-w-[80ch]'
      )}
      dangerouslySetInnerHTML={{
        __html: markdown.makeHtml(content),
      }}
    />
  </>
)
