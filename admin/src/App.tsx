import { useEffect, useState } from 'react';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Tenants from './pages/Tenants';
import TenantDetail from './pages/TenantDetail';
import Leads from './pages/Leads';
import EmbedCode from './pages/EmbedCode';
import Login from './pages/Login';
import { useAuth } from './hooks/useAuth';

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

// Reads #auth=KEY&tenant=ID&name=NAME from the URL hash, verifies the key,
// logs in, sets tenant context, and redirects to dashboard.
function AutoAuthHandler({ onLogin }: { onLogin: (key: string) => void }) {
  const navigate = useNavigate();
  const [done, setDone] = useState(false);

  useEffect(() => {
    const hash = window.location.hash.slice(1);
    if (!hash) { setDone(true); return; }
    const params = new URLSearchParams(hash);
    const authKey = params.get('auth');
    const tenantId = params.get('tenant');
    const tenantName = params.get('name') || '';
    const clientId = params.get('client') || '';
    if (!authKey) { setDone(true); return; }

    // Clear hash immediately — key must not sit in the address bar
    history.replaceState(null, '', window.location.pathname);

    fetch('/api/admin/verify', { headers: { 'X-Blueprint-Admin-Key': authKey } })
      .then(r => {
        if (r.ok) {
          onLogin(authKey);
          if (tenantId) localStorage.setItem('bp-tenant-id', tenantId);
          if (tenantName) localStorage.setItem('bp-tenant-name', decodeURIComponent(tenantName));
          // Only reset client scope when the hash explicitly came with no client
          // (i.e. this is a direct admin login, not a portal launch missing the param)
          if (clientId) {
            localStorage.setItem('bp-client-id', clientId);
          } else if (!tenantId) {
            // No tenant context at all → pure admin login, clear any stale client scope
            localStorage.removeItem('bp-client-id');
          }
          navigate('/dashboard', { replace: true });
        }
      })
      .catch(() => {})
      .finally(() => setDone(true));
  }, []);

  if (!done) return null;
  return null;
}

export default function App() {
  const { isAuthenticated, login, logout } = useAuth();

  return (
    <>
      <AutoAuthHandler onLogin={login} />
      <Routes>
        <Route
          path="/login"
          element={
            isAuthenticated
              ? <Navigate to="/dashboard" replace />
              : <Login onLogin={login} />
          }
        />
        <Route
          element={
            <RequireAuth>
              <Layout onLogout={logout} />
            </RequireAuth>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/tenants" element={<Tenants />} />
          <Route path="/tenants/:id" element={<TenantDetail />} />
          <Route path="/leads" element={<Leads />} />
          <Route path="/tenants/:id/embed" element={<EmbedCode />} />
        </Route>
      </Routes>
    </>
  );
}
