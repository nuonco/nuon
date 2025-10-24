import showdown from 'showdown'

// load extension to add target="_blank" to links
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
    filter: (text) =>
      text.replace(
        /(<table[^>]*>[\s\S]*?<\/table>)/g,
        '<div class="readme-table">$1</div>'
      ),
  },
])

// TODO(nnnat): unsure if we need variable highlighting
// load extension to highlight variables
/* showdown.extension('variables', () => {
 *   return [
 *     {
 *       type: 'html',
 *       regex: /(\{\{.+\}\})/g,
 *       replace: '<span style="font-family: monospace; color: #555f6d">$1</span>',
 *     },
 *   ]
 * }) */

// instantiate converter
const markdown = new showdown.Converter({
  extensions: ['targetlink', 'wrapTables'],
  tables: true,
  tasklists: true,
})

export const Markdown = ({ content = '' }) => (
  <div
    className="prose prose-sm dark:prose-invert !max-w-full"
    dangerouslySetInnerHTML={{
      __html: markdown.makeHtml(content),
    }}
  />
)
