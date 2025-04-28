import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function POST() {
  try {
    // Get the session token from cookies
    const cookieStore = await cookies();
    const sessionToken = cookieStore.get('session_token');

    if (sessionToken) {
      // Forward the request to your Go backend
      await fetch('http://localhost:8080/api/logout', {
        method: 'POST',
        headers: {
          Cookie: `session_token=${sessionToken.value}`,
        },
      });
    }

    // Clear the cookie
    const cookieJar = await cookies();
    cookieJar.delete('session_token'); // Make sure you're deleting the correct cookie!

    return NextResponse.json({ message: 'Logged out successfully' });
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json(
      { error: 'Logout failed' },
      { status: 500 }
    );
  }
}
