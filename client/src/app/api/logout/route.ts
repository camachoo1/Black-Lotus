import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function POST() {
  try {
    // Get cookies
    const cookieStore = await cookies();
    const csrfToken = cookieStore.get('csrf_token')?.value;

    // Create headers with CSRF token
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    // Call your backend logout endpoint
    await fetch('http://localhost:8080/api/logout', {
      method: 'POST',
      headers,
      credentials: 'include',
    });

    // Create response with cookie clearing
    const response = NextResponse.json({
      message: 'Logged out successfully',
    });

    // Clear cookies
    response.cookies.delete('access_token');
    response.cookies.delete('refresh_token');
    response.cookies.delete('csrf_token');

    return response;
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json(
      { error: 'Logout failed' },
      { status: 500 }
    );
  }
}
