export const createFileDownload = (
  content: string | Blob | ArrayBuffer,
  filename: string,
  mimeType: string = 'text/plain'
): void => {
  const blob =
    content instanceof Blob
      ? content
      : content instanceof ArrayBuffer
        ? new Blob([content], { type: mimeType })
        : new Blob([content], { type: mimeType })

  const url = window.URL.createObjectURL(blob)

  const link = document.createElement('a')

  link.href = url
  link.setAttribute('download', filename)

  document.body.appendChild(link)
  link.click()

  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

export const downloadFileOnClick = ({
  content,
  filename,
  fileType = 'toml',
  mimeType = 'text/plain',
  callback,
}: {
  content: string
  filename: string
  mimeType?: string
  fileType?: string
  callback?: () => void
}) => {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const download = filename
    ? filename?.replace(/^["'_]+|["'_]+$/g, '').trim()
    : `download.${fileType}`

  const link = document.createElement('a')
  link.href = url
  link.download = download
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)

  setTimeout(() => URL.revokeObjectURL(url), 10000)
  if (callback) callback()
}
