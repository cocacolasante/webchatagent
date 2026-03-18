import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';

interface LoginProps {
  onLogin: (key: string) => void;
}

export default function Login({ onLogin }: LoginProps) {
  const [key, setKey] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);

    // Temporarily store key so the api client picks it up for the health check
    localStorage.setItem('bp-admin-key', key);
    try {
      // Validate the key against a real admin endpoint
      await api.listTenants();
      onLogin(key);
      navigate('/dashboard', { replace: true });
    } catch {
      localStorage.removeItem('bp-admin-key');
      setError('Invalid admin key. Please try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-8 w-full max-w-sm">
        <div className="flex items-center gap-3 mb-8">
          <div className="w-10 h-10 bg-primary-500 rounded-lg flex items-center justify-center text-white font-bold">
            BP
          </div>
          <div>
            <div className="font-semibold text-gray-900">Blueprint Chat</div>
            <div className="text-xs text-gray-400">Admin Dashboard</div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Admin Key
            </label>
            <input
              type="password"
              value={key}
              onChange={e => setKey(e.target.value)}
              placeholder="Enter your admin key"
              className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              required
              autoFocus
            />
          </div>

          {error && (
            <p className="text-sm text-red-600">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading || !key}
            className="w-full bg-primary-500 hover:bg-primary-600 disabled:opacity-50 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors"
          >
            {loading ? 'Verifying...' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  );
}
