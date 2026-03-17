import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, Tenant } from '../api/client';

type Tab = 'branding' | 'business' | 'knowledge' | 'scheduler' | 'leads-config' | 'leads';

export default function TenantDetail() {
  const { id } = useParams<{ id: string }>();
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<Tab>('branding');
  const [form, setForm] = useState<Partial<Tenant>>({});
  const [saveMsg, setSaveMsg] = useState('');

  useEffect(() => {
    if (!id) return;
    api.getTenant(id)
      .then(t => {
        setTenant(t);
        setForm(t);
      })
      .finally(() => setLoading(false));
  }, [id]);

  async function save() {
    if (!id) return;
    setSaving(true);
    setSaveMsg('');
    try {
      const updated = await api.updateTenant(id, form);
      setTenant(updated);
      setSaveMsg('Saved!');
      setTimeout(() => setSaveMsg(''), 2000);
    } catch (err: unknown) {
      setSaveMsg(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <div className="p-8 text-gray-400">Loading...</div>;
  if (!tenant) return <div className="p-8 text-gray-400">Tenant not found.</div>;

  const tabs: { id: Tab; label: string }[] = [
    { id: 'branding', label: 'Branding' },
    { id: 'business', label: 'Business' },
    { id: 'knowledge', label: 'Knowledge Base' },
    { id: 'scheduler', label: 'Scheduler' },
    { id: 'leads-config', label: 'Lead Capture' },
    { id: 'leads', label: 'Leads' },
  ];

  return (
    <div className="p-8">
      <div className="flex items-center gap-3 mb-6">
        <Link to="/tenants" className="text-gray-400 hover:text-gray-600 text-sm">← Tenants</Link>
        <span className="text-gray-300">/</span>
        <h1 className="text-xl font-bold text-gray-900">{tenant.name}</h1>
        <span className={tenant.isActive ? 'badge-green' : 'badge-red'}>
          {tenant.isActive ? 'Active' : 'Inactive'}
        </span>
      </div>

      {/* Quick links */}
      <div className="flex gap-2 mb-6">
        <Link to={`/tenants/${id}/embed`} className="btn-secondary text-xs">📋 Embed Code</Link>
        <button onClick={() => navigator.clipboard.writeText(tenant.id)} className="btn-secondary text-xs">Copy ID</button>
        <button onClick={() => api.rotateKey(id!).then(r => { setTenant(t => t ? { ...t, apiKey: r.apiKey } : t); setForm(f => ({ ...f, apiKey: r.apiKey })); }).catch(alert)} className="btn-secondary text-xs">🔑 Rotate Key</button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-gray-200 mb-6 overflow-x-auto">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              activeTab === tab.id
                ? 'border-primary-500 text-primary-500'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="card p-6 max-w-2xl">
        {activeTab === 'branding' && (
          <div className="space-y-4">
            <h3 className="font-semibold mb-4">Branding & Appearance</h3>
            <div><label className="label">Bot Name</label><input className="input" value={form.botName || ''} onChange={e => setForm(p => ({ ...p, botName: e.target.value }))} /></div>
            <div><label className="label">Avatar URL</label><input className="input" value={form.avatarUrl || ''} onChange={e => setForm(p => ({ ...p, avatarUrl: e.target.value }))} placeholder="https://..." /></div>
            <div className="flex items-center gap-3">
              <div className="flex-1"><label className="label">Primary Color</label><input className="input" value={form.primaryColor || '#6C63FF'} onChange={e => setForm(p => ({ ...p, primaryColor: e.target.value }))} /></div>
              <input type="color" value={form.primaryColor || '#6C63FF'} onChange={e => setForm(p => ({ ...p, primaryColor: e.target.value }))} className="w-10 h-10 rounded border border-gray-200 cursor-pointer mt-5" />
            </div>
            <div>
              <label className="label">Greeting Message</label>
              <textarea className="input" rows={3} value={form.greeting || ''} onChange={e => setForm(p => ({ ...p, greeting: e.target.value }))} />
            </div>
            <div>
              <label className="label">Position</label>
              <select className="input" value={form.position || 'bottom-right'} onChange={e => setForm(p => ({ ...p, position: e.target.value }))}>
                <option value="bottom-right">Bottom Right</option>
                <option value="bottom-left">Bottom Left</option>
              </select>
            </div>
          </div>
        )}

        {activeTab === 'business' && (
          <div className="space-y-4">
            <h3 className="font-semibold mb-4">Business Information</h3>
            <p className="text-sm text-gray-500">Business info is stored as structured data in the Knowledge Base section. Use the Knowledge Base tab to update services, FAQs, and custom instructions.</p>
            <div>
              <label className="label">Tenant ID (read-only)</label>
              <input className="input bg-gray-50" readOnly value={tenant.id} />
            </div>
            <div>
              <label className="label">API Key (read-only)</label>
              <input className="input bg-gray-50" readOnly value={tenant.apiKey} type="password" />
            </div>
          </div>
        )}

        {activeTab === 'knowledge' && (
          <div className="space-y-4">
            <h3 className="font-semibold mb-4">Knowledge Base</h3>
            <p className="text-sm text-gray-500 mb-4">Edit the raw knowledge base JSON. Changes take effect immediately on new conversations.</p>
            <div>
              <label className="label">Knowledge Base (JSON)</label>
              <textarea
                className="input font-mono text-xs"
                rows={16}
                value={typeof form.knowledgeBase === 'object' ? JSON.stringify(form.knowledgeBase, null, 2) : '{}'}
                onChange={e => {
                  try {
                    const parsed = JSON.parse(e.target.value);
                    setForm(p => ({ ...p, knowledgeBase: parsed }));
                  } catch { /* ignore parse errors while typing */ }
                }}
              />
            </div>
          </div>
        )}

        {activeTab === 'scheduler' && (
          <div className="space-y-4">
            <h3 className="font-semibold mb-4">Scheduler Configuration</h3>
            <div>
              <label className="label">Scheduler Provider</label>
              <select className="input" value={form.schedulerType || ''} onChange={e => setForm(p => ({ ...p, schedulerType: e.target.value }))}>
                <option value="">None (disabled)</option>
                <option value="calcom">Cal.com</option>
                <option value="calendly">Calendly</option>
                <option value="google">Google Calendar</option>
              </select>
            </div>
            {form.schedulerType && (
              <p className="text-sm text-gray-500">
                Configure API credentials via the API endpoint or raw config editor. Credentials are encrypted at rest.
              </p>
            )}
          </div>
        )}

        {activeTab === 'leads-config' && (
          <div className="space-y-4">
            <h3 className="font-semibold mb-4">Lead Capture Settings</h3>
            <div className="flex items-center gap-3">
              <input type="checkbox" id="lce" checked={form.leadCaptureEnabled || false} onChange={e => setForm(p => ({ ...p, leadCaptureEnabled: e.target.checked }))} className="w-4 h-4 accent-primary-500" />
              <label htmlFor="lce" className="text-sm font-medium text-gray-700">Enable Lead Capture</label>
            </div>
            <div><label className="label">Lead Webhook URL</label><input className="input" type="url" value={(form as Record<string, unknown>).leadWebhookUrl as string || ''} onChange={e => setForm(p => ({ ...p, leadWebhookUrl: e.target.value }))} placeholder="https://..." /></div>
            <div><label className="label">Notification Email</label><input className="input" type="email" value={(form as Record<string, unknown>).leadNotifyEmail as string || ''} onChange={e => setForm(p => ({ ...p, leadNotifyEmail: e.target.value }))} /></div>
            <div><label className="label">Discord Webhook URL</label><input className="input" type="url" value={(form as Record<string, unknown>).discordWebhookUrl as string || ''} onChange={e => setForm(p => ({ ...p, discordWebhookUrl: e.target.value }))} placeholder="https://discord.com/api/webhooks/..." /></div>
          </div>
        )}

        {activeTab === 'leads' && (
          <LeadsTab tenantId={tenant.id} />
        )}

        {activeTab !== 'leads' && (
          <div className="flex items-center gap-3 mt-6 pt-6 border-t border-gray-100">
            <button className="btn-primary" onClick={save} disabled={saving}>{saving ? 'Saving...' : 'Save Changes'}</button>
            {saveMsg && <span className={`text-sm ${saveMsg === 'Saved!' ? 'text-green-600' : 'text-red-600'}`}>{saveMsg}</span>}
          </div>
        )}
      </div>
    </div>
  );
}

function LeadsTab({ tenantId }: { tenantId: string }) {
  const [leads, setLeads] = useState<Array<{ id: string; firstName: string; lastName?: string; email: string; phone?: string; status: string; createdAt: string }>>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.listLeads(tenantId)
      .then(r => setLeads(r.leads || []))
      .finally(() => setLoading(false));
  }, [tenantId]);

  return (
    <div>
      <h3 className="font-semibold mb-4">Captured Leads</h3>
      {loading ? (
        <p className="text-sm text-gray-400">Loading...</p>
      ) : leads.length === 0 ? (
        <p className="text-sm text-gray-400">No leads captured yet.</p>
      ) : (
        <div className="space-y-2">
          {leads.map(lead => (
            <div key={lead.id} className="p-3 bg-gray-50 rounded-lg text-sm">
              <div className="font-medium">{lead.firstName} {lead.lastName}</div>
              <div className="text-gray-500">{lead.email} {lead.phone ? `· ${lead.phone}` : ''}</div>
              <div className="flex items-center gap-2 mt-1">
                <span className={`badge ${lead.status === 'new' ? 'badge-blue' : 'badge-green'}`}>{lead.status}</span>
                <span className="text-gray-400 text-xs">{new Date(lead.createdAt).toLocaleDateString()}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
