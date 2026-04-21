export default {
  title: 'VCS Connections/RepoTypeFilter',
}

import { useState } from 'react'
import { RepoTypeFilter, REPO_FILTER_OPTIONS, type TRepoFilterType } from './RepoTypeFilter'

export const Default = () => {
  const [selected, setSelected] = useState<TRepoFilterType[]>(REPO_FILTER_OPTIONS)
  return <RepoTypeFilter selected={selected} onChange={setSelected} />
}

export const OneSelected = () => {
  const [selected, setSelected] = useState<TRepoFilterType[]>(['private'])
  return <RepoTypeFilter selected={selected} onChange={setSelected} />
}

export const TwoSelected = () => {
  const [selected, setSelected] = useState<TRepoFilterType[]>(['public', 'fork'])
  return <RepoTypeFilter selected={selected} onChange={setSelected} />
}
