import React, { useEffect, useState } from 'react';
import { api, Lead, Tenant } from '../api/client';

export default function Leads() {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [leads, setLeads] = useState<Lead[]>([]);
  const [selectedTenant, setSelectedTenant] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api.listTenants().then(r => setTenants(r.tenants || []));
  }, []);

  useEffect(() => {
    if (!selectedTenant) return;
    setLoading(true);
    api.listLeads(selectedTenant)
      .then(r => setLeads(r.leads || []))
      .finally(() => setLoading(false));
  }, [selectedTenant]);

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Leads</h1>

      <div className="mb-6">
        <label className="label">Filter by Tenant</label>
        <select className="input max-w-xs" value={selectedTenant} onChange={e => setSelectedTenant(e.target.value)}>
          <option value="">Select a tenant...</option>
          {tenants.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
        </select>
      </div>

      {selectedTenant && (
        <div className="card overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-100">
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Name</th>
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Email</th>
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Phone</th>
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Captured</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {loading ? (
                <tr><td colSpan={5} className="px-6 py-8 text-center text-sm text-gray-400">Loading...</td></tr>
              ) : leads.length === 0 ? (
                <tr><td colSpan={5} className="px-6 py-8 text-center text-sm text-gray-400">No leads found.</td></tr>
              ) : leads.map(l => (
                <tr key={l.id}>
                  <td className="px-6 py-3 text-sm font-medium">{l.firstName} {l.lastName}</td>
                  <td className="px-6 py-3 text-sm text-gray-600">{l.email}</td>
                  <td className="px-6 py-3 text-sm text-gray-600">{l.phone || '—'}</td>
                  <td className="px-6 py-3"><span className={`badge ${l.status === 'new' ? 'badge-blue' : 'badge-green'}`}>{l.status}</span></td>
                  <td className="px-6 py-3 text-sm text-gray-400">{new Date(l.createdAt).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
