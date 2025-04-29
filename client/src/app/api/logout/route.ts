import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function POST() {
  try {
    // Get both tokens from cookies
    const cookieStore = await cookies();
    const accessToken = cookieStore.get('access_token');
    const refreshToken = cookieStore.get('refresh_token');

    // Forward the request to your Go backend with the tokens
    if (accessToken || refreshToken) {
      const cookieHeader = [];

      if (accessToken) {
        cookieHeader.push(`access_token=${accessToken.value}`);
      }

      if (refreshToken) {
        cookieHeader.push(`refresh_token=${refreshToken.value}`);
      }

      await fetch('http://localhost:8080/api/logout', {
        method: 'POST',
        headers: {
          Cookie: cookieHeader.join('; '),
        },
      });
    }

    // Clear the cookies
    const response = NextResponse.json({
      message: 'Logged out successfully',
    });

    // Clear access token cookie
    response.cookies.delete('access_token');

    // Clear refresh token cookie
    response.cookies.delete('refresh_token');

    return response;
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json(
      { error: 'Logout failed' },
      { status: 500 }
    );
  }
}
