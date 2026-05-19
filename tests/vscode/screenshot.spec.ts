import { test, expect, Page } from "@playwright/test";

const DIAGNOSTICS_WAIT = 8000; // time for LSP to produce diagnostics

test.describe("temporal-lsp VS Code diagnostics", () => {
  let page: Page;

  test.beforeAll(async ({ browser }) => {
    page = await browser.newPage();
    // Navigate to code-server
    await page.goto("/");
    // Wait for VS Code to fully load
    await page.waitForSelector(".monaco-workbench", { timeout: 30000 });
    await page.waitForTimeout(3000);
    // Close the Welcome tab by clicking its close button
    const welcomeTab = page.locator('.tab:has-text("Welcome")');
    if (await welcomeTab.isVisible().catch(() => false)) {
      await welcomeTab.hover();
      await page.locator('.tab:has-text("Welcome") .codicon-close').click();
      await page.waitForTimeout(1000);
    }
    // Close the secondary sidebar (Chat panel) via command palette
    await page.keyboard.press("Control+Shift+p");
    await page.waitForSelector(".quick-input-widget", { timeout: 5000 });
    await page.keyboard.type("View: Close Secondary Side Bar", { delay: 20 });
    await page.waitForTimeout(500);
    await page.keyboard.press("Enter");
    await page.waitForTimeout(500);
    // Press Escape in case command wasn't found
    await page.keyboard.press("Escape");
    await page.waitForTimeout(300);
  });

  test.afterAll(async () => {
    await page.close();
  });

  async function openFile(filePath: string) {
    // Click on file in the Explorer sidebar
    const fileItem = page.locator(`.monaco-list-row:has-text("${filePath}")`).first();
    if (await fileItem.isVisible().catch(() => false)) {
      await fileItem.dblclick();
    } else {
      // Fallback: use quick open dialog (Ctrl+P)
      await page.keyboard.press("Control+p");
      await page.waitForSelector(".quick-input-widget", { timeout: 5000 });
      await page.keyboard.type(filePath, { delay: 30 });
      await page.waitForTimeout(500);
      await page.keyboard.press("Enter");
    }
    // Wait for file to open and LSP to analyze
    await page.waitForTimeout(DIAGNOSTICS_WAIT);
    // Open the Problems panel to show diagnostics list
    await page.keyboard.press("Control+Shift+m");
    await page.waitForTimeout(1000);
  }

  test("bad_time_workflow.py shows time/sleep violations", async () => {
    await openFile("bad_time_workflow.py");
    await page.screenshot({
      path: "/output/bad_time_workflow_diagnostics.png",
      fullPage: true,
    });
    await expect(page).toHaveScreenshot("bad_time_workflow.png", {
      maxDiffPixelRatio: 0.01,
    });
  });

  test("bad_env_workflow.py shows env access violations", async () => {
    await openFile("bad_env_workflow.py");
    await page.screenshot({
      path: "/output/bad_env_workflow_diagnostics.png",
      fullPage: true,
    });
    await expect(page).toHaveScreenshot("bad_env_workflow.png", {
      maxDiffPixelRatio: 0.01,
    });
  });

  test("bad_logging_workflow.py shows logging violations", async () => {
    await openFile("bad_logging_workflow.py");
    await page.screenshot({
      path: "/output/bad_logging_workflow_diagnostics.png",
      fullPage: true,
    });
    await expect(page).toHaveScreenshot("bad_logging_workflow.png", {
      maxDiffPixelRatio: 0.01,
    });
  });

  test("good_workflow.py shows no diagnostics", async () => {
    // Close all other editors so Problems panel is clean
    await page.keyboard.press("Control+k");
    await page.keyboard.press("Control+w");
    await page.waitForTimeout(500);
    await openFile("good_workflow.py");
    await page.screenshot({
      path: "/output/good_workflow_no_diagnostics.png",
      fullPage: true,
    });
    await expect(page).toHaveScreenshot("good_workflow.png", {
      maxDiffPixelRatio: 0.01,
    });
  });
});

// Single continuous video: opens all files in sequence for a demo recording
test("demo video walkthrough", async ({ page }) => {
  await page.goto("/");
  await page.waitForSelector(".monaco-workbench", { timeout: 30000 });
  await page.waitForTimeout(3000);
  // Close Welcome tab
  const welcomeTab = page.locator('.tab:has-text("Welcome")');
  if (await welcomeTab.isVisible().catch(() => false)) {
    await welcomeTab.hover();
    await page.locator('.tab:has-text("Welcome") .codicon-close').click();
    await page.waitForTimeout(1000);
  }
  // Close the secondary sidebar (Chat panel)
  await page.keyboard.press("Control+Shift+p");
  await page.waitForSelector(".quick-input-widget", { timeout: 5000 });
  await page.keyboard.type("View: Close Secondary Side Bar", { delay: 20 });
  await page.waitForTimeout(500);
  await page.keyboard.press("Enter");
  await page.waitForTimeout(500);
  await page.keyboard.press("Escape");
  await page.waitForTimeout(300);

  const files = [
    "bad_time_workflow.py",
    "bad_env_workflow.py",
    "bad_logging_workflow.py",
    "good_workflow.py",
  ];

  for (const file of files) {
    const fileItem = page.locator(`.monaco-list-row:has-text("${file}")`).first();
    if (await fileItem.isVisible().catch(() => false)) {
      await fileItem.dblclick();
    } else {
      await page.keyboard.press("Control+p");
      await page.waitForSelector(".quick-input-widget", { timeout: 5000 });
      await page.keyboard.type(file, { delay: 50 });
      await page.waitForTimeout(500);
      await page.keyboard.press("Enter");
    }
    // Wait for diagnostics to appear
    await page.waitForTimeout(DIAGNOSTICS_WAIT);
    // Open the Problems panel
    await page.keyboard.press("Control+Shift+m");
    await page.waitForTimeout(1000);
    // Pause to let viewer see the diagnostics
    await page.waitForTimeout(3000);
  }
});
