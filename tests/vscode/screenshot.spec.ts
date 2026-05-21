import { test, expect, Page } from "@playwright/test";

// Single continuous video: cycles through focused violation files
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

  // Close secondary sidebar
  await page.keyboard.press("Control+Shift+p");
  await page.waitForSelector(".quick-input-widget", { timeout: 5000 });
  await page.keyboard.type("View: Close Secondary Side Bar", { delay: 20 });
  await page.waitForTimeout(500);
  await page.keyboard.press("Enter");
  await page.waitForTimeout(500);
  await page.keyboard.press("Escape");
  await page.waitForTimeout(300);

  // Hide Explorer sidebar for cleaner view
  await page.keyboard.press("Control+b");
  await page.waitForTimeout(500);

  // Open Problems panel FIRST — stays visible for entire demo
  await page.keyboard.press("Control+Shift+m");
  await page.waitForTimeout(1000);

  // Each file shows ONE violation type in a different language
  const files = [
    "python_time.py",      // Python: datetime.now() violation
    "java_sleep.java",     // Java: Thread.sleep() violation
    "rust_io.rs",          // Rust: std::fs file read violation
    "python_random.py",    // Python: random.randint() violation
    "java_thread.java",    // Java: new Thread() violation
    "rust_mutex.rs",       // Rust: Mutex::new() violation
    "python_good.py",      // Clean workflow — no diagnostics
  ];

  for (const file of files) {
    await page.keyboard.press("Control+p");
    await page.waitForSelector(".quick-input-widget", { timeout: 5000 });
    await page.keyboard.type(file, { delay: 40 });
    await page.waitForTimeout(400);
    await page.keyboard.press("Enter");

    // Wait for LSP diagnostics to appear
    await page.waitForTimeout(3000);

    // Pause for viewer to read
    await page.waitForTimeout(2000);
  }
});
