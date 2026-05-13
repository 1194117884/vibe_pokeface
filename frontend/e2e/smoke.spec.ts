import { test, expect } from "@playwright/test";

test("Playwright is working", async ({ page }) => {
  await page.goto("/");
  // Just verify the page loads without crashing
  const title = await page.title();
  expect(title).toBeDefined();
});
