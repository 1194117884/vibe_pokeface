import { Page } from "@playwright/test";

export async function injectAuth(page: Page, userId: number = 1) {
  const header = Buffer.from(JSON.stringify({ alg: "HS256", typ: "JWT" })).toString("base64");
  const payload = Buffer.from(JSON.stringify({ user_id: userId, sub: String(userId), exp: Date.now() + 3600000 })).toString("base64");
  const token = `${header}.${payload}.fake-signature`;
  await page.evaluate((t) => localStorage.setItem("token", t), token);
}

export async function waitForConnection(page: Page) {
  await page.waitForSelector("header", { timeout: 10000 });
}

export function createHand(...cardIds: number[]): number[] {
  return cardIds;
}
