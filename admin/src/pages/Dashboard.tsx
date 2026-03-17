import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api, Tenant } from '../api/client';

export default function Dashboard() {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [health, setHealth] = useState<{ status: string } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.listTenants().then(r => setTenants(r.tenants || [])),
      api.health().then(setHealth),
    ]).finally(() => setLoading(false));
  }, []);

  const activeCount = tenants.filter(t => t.isActive).length;

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <div className="card p-6">
          <div className="text-3xl font-bold text-primary-500">{loading ? '—' : activeCount}</div>
          <div className="text-sm text-gray-500 mt-1">Active Tenants</div>
        </div>
        <div className="card p-6">
          <div className="text-3xl font-bold text-gray-800">{loading ? '—' : tenants.length}</div>
          <div className="text-sm text-gray-500 mt-1">Total Tenants</div>
        </div>
        <div className="card p-6">
          <div className={`text-3xl font-bold ${health?.status === 'ok' ? 'text-green-500' : 'text-red-500'}`}>
            {health ? (health.status === 'ok' ? '✓' : '!') : '—'}
          </div>
          <div className="text-sm text-gray-500 mt-1">System Health</div>
        </div>
      </div>

      {/* Recent tenants */}
      <div className="card">
        <div className="p-6 border-b border-gray-100 flex items-center justify-between">
          <h2 className="font-semibold text-gray-900">Recent Tenants</h2>
          <Link to="/tenants" className="btn-secondary text-xs">View all</Link>
        </div>
        <div className="divide-y divide-gray-50">
          {loading ? (
            <div className="p-6 text-sm text-gray-400">Loading...</div>
          ) : tenants.length === 0 ? (
            <div className="p-6 text-sm text-gray-400">No tenants yet. <Link to="/tenants" className="text-primary-500 underline">Create one</Link>.</div>
          ) : (
            tenants.slice(0, 5).map(t => (
              <Link key={t.id} to={`/tenants/${t.id}`} className="flex items-center justify-between p-4 hover:bg-gray-50 transition-colors">
                <div>
                  <div className="font-medium text-sm text-gray-900">{t.name}</div>
                  <div className="text-xs text-gray-400">{t.botName} · {t.plan}</div>
                </div>
                <span className={t.isActive ? 'badge-green' : 'badge-gray'}>
                  {t.isActive ? 'Active' : 'Inactive'}
                </span>
              </Link>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
