import { test, expect } from '@playwright/test'

test.describe('Auth redirects', () => {
  test('root path redirects unauthenticated users to login', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
  })

  test('protected /channels/@me redirects to login', async ({ page }) => {
    await page.goto('/channels/@me')
    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
  })

  test('protected server channel route redirects to login', async ({ page }) => {
    await page.goto('/channels/123456/789')
    await expect(page).toHaveURL(/\/login/)
  })

  test('redirect preserves intended destination in query param', async ({ page }) => {
    await page.goto('/channels/@me')
    await expect(page).toHaveURL(/redirect=.*channels.*@me/)
  })

  test('login page is accessible without auth', async ({ page }) => {
    await page.goto('/login')
    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
  })
})
