export default {
  title: 'Notebooks/NotebookCellCard',
}

import { useState } from 'react'
import { NotebookCellCard } from './NotebookCellCard'

const noop = () => {}

export const Default = () => {
  const [name, setName] = useState('list pods')
  const [script, setScript] = useState('#!/bin/bash\nkubectl get pods')

  return (
    <NotebookCellCard
      index={0}
      name={name}
      script={script}
      isDirty={false}
      isSaving={false}
      isRunning={false}
      isDeleting={false}
      onNameChange={setName}
      onScriptChange={setScript}
      onSave={noop}
      onRun={noop}
      onDelete={noop}
    />
  )
}

export const WithLastRun = () => (
  <NotebookCellCard
    index={1}
    name="echo hello"
    script={'#!/bin/bash\necho hello'}
    isDirty
    isSaving={false}
    isRunning={false}
    isDeleting={false}
    runStatus="active"
    runStatusDescription="Cell is running"
    runCreatedAt={new Date().toISOString()}
    onNameChange={noop}
    onScriptChange={noop}
    onSave={noop}
    onRun={noop}
    onDelete={noop}
  />
)
