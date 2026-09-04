export default {
  stories:
    'client/lite/components/{atoms,molecules,organisms,templates}/**/*.stories.{tsx,jsx,ts,js}',
  port: 61001,
  previewPort: 61002,
  outDir: 'build-ladle-lite',
  storyOrder: (stories) => {
    const groupOf = (id) => id.slice(0, id.lastIndexOf('--'))
    const nameOf = (id) => id.slice(id.lastIndexOf('--') + 2)
    const ordered = []
    const done = new Set()
    stories.forEach((story) => {
      const group = groupOf(story)
      if (done.has(group)) return
      done.add(group)
      const inGroup = stories.filter((id) => groupOf(id) === group)
      ordered.push(
        ...inGroup.filter((id) => nameOf(id) === 'overview'),
        ...inGroup.filter((id) => nameOf(id) !== 'overview')
      )
    })
    return ordered
  },
}
