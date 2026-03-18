import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, Tenant } from '../api/client';

const API_BASE = import.meta.env.VITE_API_BASE || 'https://chat.blueprintautomation.tech';

const platforms = [
  { id: 'html', label: 'Raw HTML', instructions: 'Paste before the closing </body> tag of your HTML file.' },
  { id: 'wordpress', label: 'WordPress', instructions: 'Go to Appearance → Theme Editor → footer.php (or use a plugin like "Insert Headers and Footers") and paste before </body>.' },
  { id: 'squarespace', label: 'Squarespace', instructions: 'Go to Settings → Advanced → Code Injection → Footer and paste the code.' },
  { id: 'shopify', label: 'Shopify', instructions: 'Go to Online Store → Themes → Edit Code → theme.liquid and paste before </body>.' },
  { id: 'webflow', label: 'Webflow', instructions: 'Go to Project Settings → Custom Code → Footer Code and paste the script.' },
];

export default function EmbedCode() {
  const { id } = useParams<{ id: string }>();
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!id) return;
    api.getTenant(id).then(setTenant).finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div className="p-8 text-gray-400">Loading...</div>;
  if (!tenant) return <div className="p-8 text-gray-400">Tenant not found.</div>;

  const embedCode = `<!-- Blueprint Chat — powered by Blueprint Automation -->\n<script src="${API_BASE}/widget.js" data-tenant-id="${tenant.id}" data-api-base="${API_BASE}" data-position="${tenant.position || 'bottom-right'}" async></script>`;

  function copy() {
    navigator.clipboard.writeText(embedCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="p-8">
      <div className="flex items-center gap-3 mb-6">
        <Link to={`/tenants/${id}`} className="text-gray-400 hover:text-gray-600 text-sm">← {tenant.name}</Link>
        <span className="text-gray-300">/</span>
        <h1 className="text-xl font-bold text-gray-900">Embed Code</h1>
      </div>

      <div className="max-w-2xl space-y-6">
        {/* Embed snippet */}
        <div className="card p-6">
          <h2 className="font-semibold mb-3">Your Embed Snippet</h2>
          <div className="relative">
            <pre className="bg-gray-900 text-green-400 rounded-lg p-4 text-sm overflow-x-auto">{embedCode}</pre>
            <button onClick={copy} className="absolute top-3 right-3 btn-secondary text-xs">
              {copied ? '✓ Copied!' : 'Copy'}
            </button>
          </div>
        </div>

        {/* Platform instructions */}
        <div className="card p-6">
          <h2 className="font-semibold mb-4">Installation Instructions</h2>
          <div className="space-y-4">
            {platforms.map(p => (
              <div key={p.id} className="border border-gray-100 rounded-lg p-4">
                <div className="font-medium text-sm mb-1">{p.label}</div>
                <div className="text-sm text-gray-500">{p.instructions}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Preview link */}
        <div className="card p-6">
          <h2 className="font-semibold mb-2">Test It</h2>
          <p className="text-sm text-gray-500 mb-3">
            Open a test page to see the widget live. Make sure the server is running.
          </p>
          <a
            href={`${API_BASE}/widget-config?tid=${tenant.id}`}
            target="_blank"
            rel="noopener noreferrer"
            className="btn-secondary text-sm"
          >
            View Widget Config →
          </a>
        </div>
      </div>
    </div>
  );
}
