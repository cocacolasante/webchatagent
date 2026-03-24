import React, { useEffect, useState } from 'react';
import { useParams, Link, useSearchParams } from 'react-router-dom';
import { api, Tenant } from '../api/client';

type Tab = 'branding' | 'business' | 'knowledge' | 'scheduler' | 'leads-config' | 'leads';

function googleOAuthURL(tenantId: string) {
  const key = localStorage.getItem('bp-admin-key') || '';
  return `/api/admin/oauth/google/start?tenant_id=${tenantId}&admin_key=${encodeURIComponent(key)}`;
}

export default function TenantDetail() {
  const { id } = useParams<{ id: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
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
        if (searchParams.get('gcal_connected') === '1') {
          setSaveMsg('Google Calendar connected!');
          setActiveTab('scheduler');
          setSearchParams({}, { replace: true });
          setTimeout(() => setSaveMsg(''), 4000);
        }
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
            <h3 className="font-semibold mb-4">Calendar Integration</h3>
            <p className="text-sm text-gray-500 mb-2">Connect a calendar so the AI agent can check availability and book appointments.</p>

            {/* Google Calendar */}
            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none">
                    <rect width="24" height="24" rx="4" fill="#4285F4"/>
                    <path d="M17 8H7a1 1 0 00-1 1v6a1 1 0 001 1h10a1 1 0 001-1V9a1 1 0 00-1-1z" fill="white"/>
                    <path d="M9 8V6a1 1 0 012 0v2M13 8V6a1 1 0 012 0v2" stroke="#4285F4" strokeWidth="1.5"/>
                    <line x1="6" y1="12" x2="18" y2="12" stroke="#4285F4" strokeWidth="1.5"/>
                  </svg>
                  <div>
                    <div className="font-medium text-sm text-gray-900">Google Calendar</div>
                    {form.schedulerType === 'google' ? (
                      <div className="text-xs text-green-600 font-medium">Connected</div>
                    ) : (
                      <div className="text-xs text-gray-500">Sign in with Google to connect</div>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {form.schedulerType === 'google' ? (
                    <>
                      <span className="inline-flex items-center gap-1 text-xs text-green-700 bg-green-50 px-2 py-1 rounded-full">
                        <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd"/></svg>
                        Connected
                      </span>
                      <button
                        className="text-xs text-gray-400 hover:text-red-500 underline"
                        onClick={() => setForm(p => ({ ...p, schedulerType: '', schedulerConfig: { ...((p as Record<string, unknown>).schedulerConfig as Record<string, unknown> || {}), apiKey: '', refreshToken: '' } as Record<string, unknown> }))}
                      >
                        Disconnect
                      </button>
                      <a
                        href={googleOAuthURL(id!)}
                        className="btn-secondary text-xs"
                        onClick={e => { e.preventDefault(); window.location.href = googleOAuthURL(id!); }}
                      >
                        Reconnect
                      </a>
                    </>
                  ) : (
                    <a
                      href={googleOAuthURL(id!)}
                      className="inline-flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 shadow-sm"
                      onClick={e => { e.preventDefault(); window.location.href = googleOAuthURL(id!); }}
                    >
                      <svg className="w-4 h-4" viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/></svg>
                      Sign in with Google
                    </a>
                  )}
                </div>
              </div>
              {form.schedulerType === 'google' && (
                <div className="mt-3 pt-3 border-t border-gray-100">
                  <label className="label text-xs">Calendar ID</label>
                  <input
                    className="input text-sm"
                    placeholder="primary"
                    value={(form as Record<string, unknown>).schedulerConfig ? ((form as Record<string, unknown>).schedulerConfig as Record<string, unknown>).calendarId as string || '' : ''}
                    onChange={e => setForm(p => ({ ...p, schedulerConfig: { ...((p as Record<string, unknown>).schedulerConfig as Record<string, unknown> || {}), calendarId: e.target.value } as Record<string, unknown> }))}
                  />
                  <p className="text-xs text-gray-400 mt-1">Leave as "primary" to use the connected account's main calendar, or enter a specific calendar ID.</p>
                </div>
              )}
            </div>

            {/* Cal.com — coming soon */}
            <div className="border border-gray-100 rounded-lg p-4 opacity-50 cursor-not-allowed">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded bg-gray-200 flex items-center justify-center text-gray-400 text-xs font-bold">C</div>
                  <div>
                    <div className="font-medium text-sm text-gray-400">Cal.com <span className="text-xs font-normal text-gray-400">(coming soon)</span></div>
                    <div className="text-xs text-gray-400">API key integration</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Calendly — coming soon */}
            <div className="border border-gray-100 rounded-lg p-4 opacity-50 cursor-not-allowed">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded bg-gray-200 flex items-center justify-center text-gray-400 text-xs font-bold">Cy</div>
                  <div>
                    <div className="font-medium text-sm text-gray-400">Calendly <span className="text-xs font-normal text-gray-400">(coming soon)</span></div>
                    <div className="text-xs text-gray-400">API key integration</div>
                  </div>
                </div>
              </div>
            </div>
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

        {activeTab !== 'leads' && !(activeTab === 'scheduler' && form.schedulerType !== 'google') && (
          <div className="flex items-center gap-3 mt-6 pt-6 border-t border-gray-100">
            <button className="btn-primary" onClick={save} disabled={saving}>{saving ? 'Saving...' : 'Save Changes'}</button>
            {saveMsg && <span className={`text-sm ${saveMsg.includes('connected') || saveMsg === 'Saved!' ? 'text-green-600' : 'text-red-600'}`}>{saveMsg}</span>}
          </div>
        )}
        {activeTab === 'scheduler' && form.schedulerType !== 'google' && saveMsg && (
          <div className="mt-4">
            <span className={`text-sm ${saveMsg.includes('connected') ? 'text-green-600' : 'text-red-600'}`}>{saveMsg}</span>
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
