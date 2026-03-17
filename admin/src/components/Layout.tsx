import React, { useState } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';

function NavItem({ to, icon, label }: { to: string; icon: string; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
          isActive
            ? 'bg-primary-500 text-white'
            : 'text-gray-600 hover:bg-gray-100'
        }`
      }
    >
      <span className="text-lg">{icon}</span>
      {label}
    </NavLink>
  );
}

export default function Layout() {
  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-60 bg-white border-r border-gray-100 flex flex-col fixed h-full">
        <div className="p-6 border-b border-gray-100">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-primary-500 rounded-lg flex items-center justify-center text-white font-bold text-sm">
              BP
            </div>
            <div>
              <div className="font-semibold text-sm text-gray-900">Blueprint Chat</div>
              <div className="text-xs text-gray-400">Admin Dashboard</div>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          <NavItem to="/dashboard" icon="📊" label="Dashboard" />
          <NavItem to="/tenants" icon="🏢" label="Tenants" />
          <NavItem to="/leads" icon="👤" label="All Leads" />
        </nav>

        <div className="p-4 border-t border-gray-100">
          <div className="text-xs text-gray-400 text-center">
            Blueprint Automation v1.1
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 ml-60 min-h-screen">
        <Outlet />
      </main>
    </div>
  );
}
