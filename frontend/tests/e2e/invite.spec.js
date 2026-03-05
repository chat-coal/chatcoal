import { test, expect } from '@playwright/test'

test.describe('Invite page', () => {
  test('shows invalid invite error when API returns an error', async ({ page }) => {
    await page.route('**/api/invites/**', (route) =>
      route.fulfill({ status: 404, contentType: 'application/json', body: JSON.stringify({ error: 'Not found' }) }),
    )

    await page.goto('/invite/bad-code')

    await expect(page.getByRole('heading', { name: 'Invalid Invite' })).toBeVisible()
    await expect(page.getByText('This invite is invalid or has expired.')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Go Home' })).toBeVisible()
  })

  test('shows server info and accept button when invite is valid', async ({ page }) => {
    await page.route('**/api/invites/**', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ server: { id: '111', name: 'Cool Community' } }),
      }),
    )

    await page.goto('/invite/good-code')

    await expect(page.getByText("You've been invited to join")).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Cool Community' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Accept Invite' })).toBeVisible()
  })

  test('shows loading state before API responds', async ({ page }) => {
    let resolveRoute
    const routePromise = new Promise((resolve) => { resolveRoute = resolve })

    await page.route('**/api/invites/**', async (route) => {
      await routePromise
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ server: { id: '222', name: 'Slow Server' } }),
      })
    })

    await page.goto('/invite/slow-code')
    await expect(page.getByText('Loading invite...')).toBeVisible()

    resolveRoute()
    await expect(page.getByRole('heading', { name: 'Slow Server' })).toBeVisible()
  })
})
