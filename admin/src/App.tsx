import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
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

export default function App() {
  const { isAuthenticated, login, logout } = useAuth();

  return (
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
  );
}
