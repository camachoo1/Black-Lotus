import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const returnTo = url.searchParams.get('returnTo') || '/';

    // Get the GitHub auth URL from your backend
    const response = await fetch(
      `http://localhost:8080/api/auth/github?returnTo=${encodeURIComponent(
        returnTo
      )}`
    );

    if (!response.ok) {
      throw new Error('Failed to get GitHub auth URL');
    }

    const data = await response.json();

    // Directly redirect the user to GitHub instead of returning JSON
    return NextResponse.redirect(data.url);
  } catch (error) {
    console.error('GitHub auth error:', error);
    // If something goes wrong, redirect to an error page
    return NextResponse.redirect(new URL('/auth/error', request.url));
  }
}
