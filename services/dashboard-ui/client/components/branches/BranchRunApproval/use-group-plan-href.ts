export const getGroupName = (name?: string) =>
  name?.replace(/^plan install group:\s*/i, '').trim() || 'install group'
