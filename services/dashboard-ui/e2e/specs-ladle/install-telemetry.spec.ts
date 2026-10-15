import { expect, test } from '@playwright/test'

test('a bootstrap credential can be copied before closing the dialog', async ({
  page,
  context,
}) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write'])
  await page.goto(
    '/?story=features--installs--collector-credential--default&mode=preview',
    { waitUntil: 'domcontentloaded' }
  )
  await page.getByRole('button', { name: 'Open modal' }).click()
  const dialog = page.getByRole('dialog')
  await dialog.getByRole('button', { name: 'Create credential' }).click()
  await dialog.getByRole('button', { name: 'Copy bootstrap token' }).click()
  expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(
    'example-not-a-real-token'
  )
  await dialog.getByRole('button', { name: 'Done', exact: true }).click()
  await expect(dialog).toHaveCount(0)
})
