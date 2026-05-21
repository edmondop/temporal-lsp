import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: ".",
  timeout: 120000,
  snapshotDir: "/output/snapshots",
  use: {
    baseURL: "http://localhost:8080",
    viewport: { width: 1400, height: 900 },
    deviceScaleFactor: 2,
    video: {
      mode: "on",
      size: { width: 1400, height: 900 },
    },
    screenshot: "on",
  },
  outputDir: "/output",
  expect: {
    toHaveScreenshot: {
      maxDiffPixelRatio: 0.01,
    },
  },
});
