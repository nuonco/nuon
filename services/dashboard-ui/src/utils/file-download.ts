export const createFileDownload = (
  content: string | Blob | ArrayBuffer,
  filename: string,
  mimeType: string = 'text/plain'
): void => {
  // Create a blob with the content and specified MIME type
  const blob =
    content instanceof Blob
      ? content
      : content instanceof ArrayBuffer
        ? new Blob([content], { type: mimeType })
        : new Blob([content], { type: mimeType })

  // Create a URL for the blob
  const url = window.URL.createObjectURL(blob)

  // Create an anchor element
  const link = document.createElement('a')

  // Set the download attributes
  link.href = url
  link.setAttribute('download', filename)

  // Append to the document, trigger the download, and clean up
  document.body.appendChild(link)
  link.click()

  // Clean up: remove the element and revoke the object URL
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}
