import showdown from 'showdown'

// load extension to add target="_blank" to links
showdown.extension('targetlink', () => {
  return [
    {
      type: 'html',
      regex: /(<a [^>]+?)(>.*<\/a>)/g,
      replace: '$1 target="_blank"$2',
    },
  ]
})


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
  extensions: ['targetlink'],
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
