import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function GET() {
  try {
    // Get all cookies
    const cookieStore = await cookies();
    const sessionToken = cookieStore.get('session_token');

    if (!sessionToken) {
      return NextResponse.json(
        { error: 'Not authenticated' },
        { status: 401 }
      );
    }

    // Forward the request to your Go backend with the session token
    const response = await fetch(
      'http://localhost:8080/api/profile',
      {
        method: 'GET',
        headers: {
          Cookie: `session_token=${sessionToken.value}`,
        },
        credentials: 'include'
      }
    );

    if (!response.ok) {
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
