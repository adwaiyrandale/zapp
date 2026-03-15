import { useState } from 'react';
import { CreditCard, Building2, DollarSign, Activity, Menu, X, Wallet } from 'lucide-react';
import { Button } from './components/ui/button';
import { PaymentsPage } from './pages/PaymentsPage';
import { SettlementsPage } from './pages/SettlementsPage';
import { LedgerPage } from './pages/LedgerPage';
import { AccountsPage } from './pages/AccountsPage';

type Page = 'payments' | 'settlements' | 'ledger' | 'accounts';

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('accounts');
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const navItems = [
    { id: 'accounts' as Page, label: 'Accounts', icon: Wallet },
    { id: 'payments' as Page, label: 'Payments', icon: CreditCard },
    { id: 'settlements' as Page, label: 'Settlements', icon: DollarSign },
    { id: 'ledger' as Page, label: 'Ledger', icon: Building2 },
  ];

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside 
        className={`${sidebarOpen ? 'w-64' : 'w-16'} bg-slate-900 text-white transition-all duration-300 flex flex-col`}
      >
        <div className="p-4 flex items-center justify-between border-b border-slate-800">
          {sidebarOpen && (
            <div className="flex items-center gap-2">
              <Activity className="h-6 w-6 text-yellow-400" />
              <span className="font-bold text-xl">Zapp</span>
            </div>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="text-slate-400 hover:text-white"
          >
            {sidebarOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </Button>
        </div>

        <nav className="flex-1 p-2 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.id}
                onClick={() => setCurrentPage(item.id)}
                className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                  currentPage === item.id
                    ? 'bg-yellow-400 text-slate-900 font-medium'
                    : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                }`}
              >
                <Icon className="h-5 w-5" />
                {sidebarOpen && <span>{item.label}</span>}
              </button>
            );
          })}
        </nav>

        <div className="p-4 border-t border-slate-800">
          {sidebarOpen && (
            <div className="text-xs text-slate-500">
              <p>Zapp v1.0.0</p>
              <p>Payment Gateway</p>
            </div>
          )}
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        <div className="p-8 bg-slate-50 min-h-screen">
          {currentPage === 'accounts' && <AccountsPage />}
          {currentPage === 'payments' && <PaymentsPage />}
          {currentPage === 'settlements' && <SettlementsPage />}
          {currentPage === 'ledger' && <LedgerPage />}
        </div>
      </main>
    </div>
  );
}

export default App;
