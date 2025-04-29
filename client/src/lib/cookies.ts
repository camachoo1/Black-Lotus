/**
 * Parses a Set-Cookie header string that may contain multiple cookies
 * Returns an array of individual cookie strings
 */

export function parseSetCookieString(
  setCookieHeader: string
): string[] {
  const cookies: string[] = [];

  let currentCookie = '';
  let withinValue = false;

  for (let i = 0; i < setCookieHeader.length; i++) {
    const char = setCookieHeader[i];

    // Toggle the withinValue flag when we encounter quotes
    if (char === '"') {
      withinValue = !withinValue;
    }

    // If we encounter a comma and we're not within a quoted value
    if (char === ',' && !withinValue) {
      // End of a cookie
      if (currentCookie.trim()) {
        cookies.push(currentCookie.trim());
      }
      currentCookie = '';
      continue;
    }

    // Otherwise, add the character to the current cookie
    currentCookie += char;
  }

  // Add the last cookie if there's anything left
  if (currentCookie.trim()) {
    cookies.push(currentCookie.trim());
  }

  return cookies;
}
