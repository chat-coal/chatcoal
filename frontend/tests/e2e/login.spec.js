import { test, expect } from '@playwright/test'

test.describe('Login page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    // Wait for Firebase auth to initialize and the page to render
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
  })

  test('renders branding and heading', async ({ page }) => {
    await expect(page.getByText('chatcoal').first()).toBeVisible()
    await expect(page.getByText('Sign in or create an account to get started')).toBeVisible()
  })

  test('renders all auth options', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Continue with Google/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /Continue with Email/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /Continue as Guest/ })).toBeVisible()
  })

  test('shows email form when clicking Continue with Email', async ({ page }) => {
    await page.getByRole('button', { name: /Continue with Email/ }).click()

    await expect(page.getByPlaceholder('Email address')).toBeVisible()
    await expect(page.getByPlaceholder('Password')).toBeVisible()
    // The toggle button hides itself once the form is shown
    await expect(page.getByRole('button', { name: /Continue with Email/ })).not.toBeVisible()
  })

  test('email form starts in sign in mode', async ({ page }) => {
    await page.getByRole('button', { name: /Continue with Email/ }).click()

    const submitBtn = page.locator('button[type="submit"]')
    await expect(submitBtn).toContainText('Sign in')
    await expect(page.getByPlaceholder('Password')).toHaveAttribute('autocomplete', 'current-password')
  })

  test('switching to Create account tab updates form', async ({ page }) => {
    await page.getByRole('button', { name: /Continue with Email/ }).click()

    // Click the "Create account" tab (inside the tab bar, not the submit button)
    const tabBar = page.locator('.flex.bg-\\[var\\(--surface-2\\)\\].rounded-xl')
    await tabBar.getByRole('button', { name: 'Create account' }).click()

    const submitBtn = page.locator('button[type="submit"]')
    await expect(submitBtn).toContainText('Create account')
    await expect(page.getByPlaceholder('Password')).toHaveAttribute('autocomplete', 'new-password')
  })

  test('shows error message on invalid email auth', async ({ page }) => {
    await page.getByRole('button', { name: /Continue with Email/ }).click()

    await page.getByPlaceholder('Email address').fill('notareal@example.com')
    await page.getByPlaceholder('Password').fill('wrongpassword')
    await page.locator('button[type="submit"]').click()

    // Firebase will return an error; we expect an error message to appear
    await expect(page.locator('p').filter({ hasText: /No account found|Invalid email or password|Something went wrong/ })).toBeVisible({ timeout: 15000 })
  })
})
