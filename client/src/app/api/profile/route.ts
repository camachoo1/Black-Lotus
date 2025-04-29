import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function GET() {
  try {
    // Get both tokens from cookies
    const cookieStore = await cookies();
    const accessToken = cookieStore.get('access_token');
    const refreshToken = cookieStore.get('refresh_token');

    // Check for authentication
    if (!accessToken) {
      if (!refreshToken) {
        return NextResponse.json(
          { error: 'Not authenticated' },
          { status: 401 }
        );
      }

      // Has refresh token but no access token
      return NextResponse.json(
        { error: 'Access token expired', code: 'token_expired' },
        { status: 401 }
      );
    }

    // Forward the request to Go backend with the access token
    const response = await fetch(
      'http://localhost:8080/api/profile',
      {
        method: 'GET',
        headers: {
          Cookie: `access_token=${accessToken.value}`,
        },
      }
    );

    if (!response.ok) {
      if (response.status === 401) {
        // Get more details to check if it's a token expiration
        try {
          const errorData = await response.json();
          if (
            errorData.code === 'token_expired' ||
            errorData.code === 'token_invalid'
          ) {
            return NextResponse.json(errorData, { status: 401 });
          }
        } catch {
          // If we can't parse JSON, just return the normal error
        }
      }

      return NextResponse.json(
        { error: `Failed to fetch profile: ${response.status}` },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Profile fetch error:', error);
    return NextResponse.json(
      { error: 'Failed to fetch profile' },
      { status: 500 }
    );
  }
}
