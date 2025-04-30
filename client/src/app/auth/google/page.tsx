'use client';
import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';

// API base URL for direct backend communication
const API_BASE =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function GoogleAuth() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { isLoading } = useAuth();

  useEffect(() => {
    if (isLoading) return;

    console.log('Fetching auth url');

    const fetchAuthUrl = async () => {
      try {
        const returnTo = searchParams.get('returnTo') || '/';

        // Get the Google auth URL directly from the Go backend
        const response = await fetch(
          `${API_BASE}/api/auth/google?returnTo=${encodeURIComponent(
            returnTo
          )}`,
          {
            credentials: 'include',
          }
        );

        const data = await response.json();

        // Redirect to Google
        if (data.url) {
          console.log('data.url:', data.url);
          window.location.href = data.url;
        } else {
          console.error('No Google authorization URL returned');
        }
      } catch (error) {
        console.error('Failed to get Google auth URL:', error);
      }
    };

    fetchAuthUrl();
  }, [searchParams, router, isLoading]);

  return (
    <div className='flex flex-col items-center justify-center min-h-screen'>
      <div className='text-center'>
        <h1 className='text-2xl font-bold mb-4'>
          Redirecting to Google...
        </h1>
        <p className='text-gray-600 mb-4'>
          Please wait while we redirect you to Google for
          authentication.
        </p>
        <div className='animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto'></div>
      </div>
    </div>
  );
}
