import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function POST() {
  try {
    // Get refresh token
    const cookieStore = await cookies();
    const refreshToken = cookieStore.get('refresh_token');

    if (!refreshToken) {
      return NextResponse.json(
        { error: 'No refresh token' },
        { status: 401 }
      );
    }

    // Forward to backend refresh endpoint
    const response = await fetch(
      'http://localhost:8080/api/refresh-token',
      {
        method: 'POST',
        headers: {
          Cookie: `refresh_token=${refreshToken.value}`,
        },
      }
    );

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to refresh token' },
        { status: response.status }
      );
    }

    // Forward the Set-Cookie header for the new access token
    const nextResponse = NextResponse.json(
      { message: 'Token refreshed successfully' },
      { status: 200 }
    );

    // Forward all Set-Cookie headers (access and refresh tokens)
    response.headers.forEach((value, key) => {
      // Only copy the Set-Cookie header
      if (key.toLowerCase() === 'set-cookie') {
        nextResponse.headers.set(key, value);
      }
    });

    return nextResponse;
  } catch (error) {
    console.error('Token refresh error:', error);
    return NextResponse.json(
      { error: 'Failed to refresh token' },
      { status: 500 }
    );
  }
}
