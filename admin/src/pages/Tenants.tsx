import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api, Tenant, ProvisionResult } from '../api/client';

export default function Tenants() {
  const clientId = localStorage.getItem('bp-client-id');
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [newTenant, setNewTenant] = useState({ name: '', botName: 'Assistant', primaryColor: '#6C63FF', greeting: 'Hi! How can I help you today?' });
  const [embedResult, setEmbedResult] = useState<ProvisionResult | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Tenant | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    api.listTenants()
      .then(r => setTenants(r.tenants || []))
      .finally(() => setLoading(false));
  }, []);

  async function handleDelete() {
    if (!deleteTarget) return;
    try {
      await api.deleteTenant(deleteTarget.id);
      setTenants(prev => prev.filter(t => t.id !== deleteTarget.id));
      setDeleteTarget(null);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Failed to delete');
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    try {
      const result = await api.createTenant({
        ...newTenant,
        ...(clientId ? { client_id: clientId } : {}),
        leadCaptureEnabled: true,
        leadFormConfig: { fields: ['name', 'email', 'phone'], requiredFields: ['name', 'email'], triggerMessage: "I'd love to connect you with our team!" },
      });
      setEmbedResult(result);
      setTenants(prev => [result.tenant, ...prev]);
      setShowForm(false);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Failed to create tenant');
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Tenants</h1>
        <button className="btn-primary" onClick={() => setShowForm(true)}>+ New Tenant</button>
      </div>

      {/* Embed code result */}
      {embedResult && (
        <div className="card p-6 mb-6 border-green-200 bg-green-50">
          <div className="flex items-start justify-between">
            <div>
              <h3 className="font-semibold text-green-800 mb-1">✓ Tenant Created: {embedResult.tenant.name}</h3>
              <p className="text-sm text-green-700 mb-3">Embed code for the new client:</p>
              <pre className="bg-white border border-green-200 rounded-lg p-3 text-xs overflow-x-auto">{embedResult.embedCode}</pre>
            </div>
            <button onClick={() => setEmbedResult(null)} className="text-green-600 hover:text-green-800 ml-4">✕</button>
          </div>
          <div className="mt-3 flex gap-2">
            <button className="btn-secondary text-xs" onClick={() => navigator.clipboard.writeText(embedResult.embedCode)}>
              Copy Embed Code
            </button>
            <Link to={`/tenants/${embedResult.tenant.id}`} className="btn-primary text-xs">Configure Tenant →</Link>
          </div>
        </div>
      )}

      {/* Create form */}
      {showForm && (
        <div className="card p-6 mb-6">
          <h3 className="font-semibold mb-4">New Tenant</h3>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label">Business Name *</label>
                <input className="input" required value={newTenant.name} onChange={e => setNewTenant(p => ({ ...p, name: e.target.value }))} placeholder="Acme Corp" />
              </div>
              <div>
                <label className="label">Bot Name</label>
                <input className="input" value={newTenant.botName} onChange={e => setNewTenant(p => ({ ...p, botName: e.target.value }))} placeholder="Assistant" />
              </div>
            </div>
            <div>
              <label className="label">Greeting</label>
              <input className="input" value={newTenant.greeting} onChange={e => setNewTenant(p => ({ ...p, greeting: e.target.value }))} />
            </div>
            <div className="flex items-center gap-3">
              <label className="label mb-0">Primary Color</label>
              <input type="color" value={newTenant.primaryColor} onChange={e => setNewTenant(p => ({ ...p, primaryColor: e.target.value }))} className="w-10 h-8 rounded cursor-pointer border border-gray-200" />
              <span className="text-sm text-gray-500">{newTenant.primaryColor}</span>
            </div>
            <div className="flex gap-2">
              <button type="submit" className="btn-primary" disabled={creating}>{creating ? 'Creating...' : 'Create Tenant'}</button>
              <button type="button" className="btn-secondary" onClick={() => setShowForm(false)}>Cancel</button>
            </div>
          </form>
        </div>
      )}

      {/* Tenants table */}
      <div className="card overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-100">
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Bot</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Plan</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Created</th>
              <th className="px-6 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {loading ? (
              <tr><td colSpan={6} className="px-6 py-8 text-center text-sm text-gray-400">Loading...</td></tr>
            ) : tenants.length === 0 ? (
              <tr><td colSpan={6} className="px-6 py-8 text-center text-sm text-gray-400">No tenants yet.</td></tr>
            ) : tenants.map(t => (
              <tr key={t.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 font-medium text-sm text-gray-900">{t.name}</td>
                <td className="px-6 py-4 text-sm text-gray-600">{t.botName}</td>
                <td className="px-6 py-4"><span className="badge-blue">{t.plan}</span></td>
                <td className="px-6 py-4"><span className={t.isActive ? 'badge-green' : 'badge-red'}>{t.isActive ? 'Active' : 'Inactive'}</span></td>
                <td className="px-6 py-4 text-sm text-gray-400">{new Date(t.createdAt).toLocaleDateString()}</td>
                <td className="px-6 py-4 text-right">
                  <Link to={`/tenants/${t.id}`} className="text-primary-500 hover:text-primary-600 text-sm font-medium">Configure →</Link>
                  <button
                    onClick={() => setDeleteTarget(t)}
                    className="text-red-500 hover:text-red-700 text-sm font-medium ml-3"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {deleteTarget && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-sm p-6">
            <h3 className="text-base font-semibold text-gray-900 mb-2">Delete Tenant</h3>
            <p className="text-sm text-gray-600 mb-5">
              Are you sure you want to delete <strong>{deleteTarget.name}</strong>? This cannot be undone.
            </p>
            <div className="flex gap-2">
              <button onClick={() => setDeleteTarget(null)} className="flex-1 px-4 py-2 text-sm text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200">
                Cancel
              </button>
              <button onClick={handleDelete} className="flex-1 px-4 py-2 text-sm text-white bg-red-600 rounded-lg hover:bg-red-700">
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
