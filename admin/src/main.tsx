import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';

// Read portal deep-link hash params BEFORE React renders.
// Portal sends: /admin/#auth=KEY&tenant=ID&name=NAME
;(function () {
  const hash = window.location.hash.slice(1);
  if (!hash) return;
  const params = new URLSearchParams(hash);
  const auth = params.get('auth');
  const tenant = params.get('tenant');
  const name = params.get('name');
  const client = params.get('client');
  if (auth) {
    localStorage.setItem('bp-admin-key', auth);
    if (tenant) localStorage.setItem('bp-tenant-id', tenant);
    if (name) localStorage.setItem('bp-tenant-name', decodeURIComponent(name));
    if (client) localStorage.setItem('bp-client-id', client);
    window.history.replaceState(null, '', '/admin/');
  }
})();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter basename="/admin">
      <App />
    </BrowserRouter>
  </React.StrictMode>
);
