const API_BASE = import.meta.env.VITE_API_BASE || '';

function headers(): HeadersInit {
  const key = import.meta.env.VITE_ADMIN_KEY || localStorage.getItem('bp-admin-key') || '';
  const h: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Blueprint-Admin-Key': key,
  };
  const clientId = localStorage.getItem('bp-client-id');
  if (clientId) h['X-Blueprint-Client-ID'] = clientId;
  return h;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const resp = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { ...headers(), ...(options.headers || {}) },
  });

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: resp.statusText }));
    throw new Error(err.error || `Request failed: ${resp.status}`);
  }

  if (resp.status === 204) return {} as T;
  return resp.json();
}

export interface Tenant {
  id: string;
  name: string;
  apiKey: string;
  plan: string;
  isActive: boolean;
  botName: string;
  avatarUrl?: string;
  primaryColor: string;
  greeting: string;
  position: string;
  schedulerType: string;
  leadCaptureEnabled: boolean;
  leadWebhookUrl?: string;
  leadNotifyEmail?: string;
  discordWebhookUrl?: string;
  businessInfo: Record<string, unknown>;
  knowledgeBase: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface Lead {
  id: string;
  tenantId: string;
  firstName: string;
  lastName?: string;
  email: string;
  phone?: string;
  sourceUrl?: string;
  status: string;
  sessionSummary?: string;
  createdAt: string;
}

export interface ProvisionResult {
  tenant: Tenant;
  embedCode: string;
}

export const api = {
  // Tenants
  listTenants: () => request<{ tenants: Tenant[]; total: number }>('/api/admin/tenants'),
  getTenant: (id: string) => request<Tenant>(`/api/admin/tenants/${id}`),
  createTenant: (data: unknown) =>
    request<ProvisionResult>('/api/admin/tenants', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateTenant: (id: string, data: unknown) =>
    request<Tenant>(`/api/admin/tenants/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteTenant: (id: string) =>
    request<void>(`/api/admin/tenants/${id}`, { method: 'DELETE' }),
  rotateKey: (id: string) =>
    request<{ apiKey: string }>(`/api/admin/tenants/${id}/rotate-key`, { method: 'POST' }),
  updateKnowledge: (id: string, data: unknown) =>
    request<Tenant>(`/api/admin/tenants/${id}/knowledge`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Leads
  listLeads: (tenantId: string) =>
    request<{ leads: Lead[]; total: number }>(`/api/admin/tenants/${tenantId}/leads`),

  // Health
  health: () => request<{ status: string }>('/api/health'),
};
