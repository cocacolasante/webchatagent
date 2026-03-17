import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Tenants from './pages/Tenants';
import TenantDetail from './pages/TenantDetail';
import Leads from './pages/Leads';
import EmbedCode from './pages/EmbedCode';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
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
