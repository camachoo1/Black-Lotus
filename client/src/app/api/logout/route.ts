import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function POST() {
  try {
    // Get the access_token and refresh_token from cookies
    const cookieStore = await cookies();
    const accessToken = cookieStore.get('access_token');
    const refreshToken = cookieStore.get('refresh_token');
    const csrfToken = cookieStore.get('csrf_token');

    // Prepare headers for backend request
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    // Add CSRF token if available
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken.value;
    }

    // Prepare cookie header
    const cookieHeader: string[] = [];
    if (accessToken)
      cookieHeader.push(`access_token=${accessToken.value}`);
    if (refreshToken)
      cookieHeader.push(`refresh_token=${refreshToken.value}`);
    if (cookieHeader.length > 0) {
      headers['Cookie'] = cookieHeader.join('; ');
    }

    // Forward the request to the backend
    const backendRes = await fetch(
      'http://localhost:8080/api/logout',
      {
        method: 'POST',
        headers,
      }
    );

    if (!backendRes.ok) {
      console.error(
        'Backend Logout Failed:',
        await backendRes.text()
      );
      return NextResponse.json({
        error: 'Logout failed on the backend',
      });
    }

    // Create the response
    const nextResponse = NextResponse.json(
      { message: 'Logged out successfully' },
      { status: 200 }
    );

    // Clear cookies on the frontend
    nextResponse.cookies.delete('access_token');
    nextResponse.cookies.delete('refresh_token');
    nextResponse.cookies.delete('csrf_token');

    return nextResponse;
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json(
      { error: 'Logout failed' },
      { status: 500 }
    );
  }
}
