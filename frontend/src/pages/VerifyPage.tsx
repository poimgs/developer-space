import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function VerifyPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { refresh } = useAuth();
  const [error, setError] = useState('');

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      setError('Missing verification token.');
      return;
    }

    // Call the backend verify endpoint to set the session cookie.
    // The endpoint returns 302 which fetch follows automatically,
    // and credentials: 'include' ensures the Set-Cookie is applied.
    fetch(`/api/auth/verify?token=${encodeURIComponent(token)}`, {
      credentials: 'include',
    })
      .then(() => refresh())
      .then(() => {
        navigate('/', { replace: true });
      })
      .catch(() => {
        setError('Invalid or expired token.');
      });
  }, [searchParams, navigate, refresh]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-stone-900 dark:text-stone-100">Verification failed</h1>
          <p className="mt-2 text-sm text-stone-600 dark:text-stone-400">{error}</p>
          <a
            href="/login"
            className="mt-4 inline-block text-sm font-medium text-amber-600 hover:text-amber-500 dark:text-amber-400"
          >
            Back to login
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-amber-600 border-t-transparent" />
    </div>
  );
}
