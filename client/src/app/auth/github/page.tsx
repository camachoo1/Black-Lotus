'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

export default function GitHubAuth() {
  const searchParams = useSearchParams();
  const router = useRouter()

  useEffect(() => {
    console.log('Fetching auth url')
    const fetchAuthUrl = async () => {
      try {
        const returnTo = searchParams.get('returnTo') || '/';

        // Get the GitHub auth URL
        const response = await fetch(
          `/api/auth/github?returnTo=${encodeURIComponent(returnTo)}`
        );
        const data = await response.json();

        // Redirect to GitHub
        if (data.url) {
          console.log('data.url:', data.url)
          router.push(data.url)
        } else {
          console.error('No GitHub authorization URL returned');
        }
      } catch (error) {
        console.error('Failed to get GitHub auth URL:', error);
      }
    };

    fetchAuthUrl();
  }, [searchParams, router]);

  return (
    <div className='flex flex-col items-center justify-center min-h-screen'>
      <div className='text-center'>
        <h1 className='text-2xl font-bold mb-4'>
          Redirecting to GitHub...
        </h1>
        <p className='text-gray-600 mb-4'>
          Please wait while we redirect you to GitHub for
          authentication.
        </p>
        <div className='animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto'></div>
      </div>
    </div>
  );
}
