import { expect, test } from '@playwright/test'

const STORY = '/?story=common--searchinput--empty&mode=preview'

test.describe('SearchInput behavior', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    await expect(page.getByPlaceholder('Search...')).toBeVisible()
  })

  test('leading spaces never enter the query', async ({ page }) => {
    const search = page.getByPlaceholder('Search...')

    await search.click()
    await page.keyboard.type('  nuon-canary')
    await expect(search).toHaveValue('nuon-canary')

    await page.keyboard.type(' prod')
    await expect(search).toHaveValue('nuon-canary prod')
  })

  test('a query of only spaces stays empty', async ({ page }) => {
    const search = page.getByPlaceholder('Search...')

    await search.click()
    await page.keyboard.type('   ')
    await expect(search).toHaveValue('')
    await expect(page.getByTitle('clear search')).toBeHidden()
  })
})
