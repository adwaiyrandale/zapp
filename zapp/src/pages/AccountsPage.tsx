import { useState, useEffect } from 'react';
import { ledgerApi, paymentApi, settlementApi } from '../api/client';
import type { Account, Payment, Settlement } from '../types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Badge } from '../components/ui/badge';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '../components/ui/table';
import { 
  RefreshCw, 
  Wallet, 
  TrendingUp, 
  TrendingDown,
  ArrowDownLeft,
  ArrowUpRight,
  ChevronDown,
  ChevronRight,
  Banknote,
  Building2,
  PieChart,
  Eye,
  EyeOff
} from 'lucide-react';

type AccountType = 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';

interface AccountGroup {
  type: AccountType;
  accounts: Account[];
  totalBalance: number;
}

interface HoldingsSummary {
  totalAssets: number;
  totalLiabilities: number;
  totalEquity: number;
  netPosition: number;
}

const ACCOUNT_TYPE_CONFIG: Record<AccountType, { label: string; icon: React.ElementType; color: string }> = {
  ASSET: { label: 'Assets', icon: Banknote, color: 'text-yellow-400' },
  LIABILITY: { label: 'Liabilities', icon: Building2, color: 'text-red-400' },
  EQUITY: { label: 'Equity', icon: PieChart, color: 'text-blue-400' },
  REVENUE: { label: 'Revenue', icon: TrendingUp, color: 'text-green-400' },
  EXPENSE: { label: 'Expenses', icon: TrendingDown, color: 'text-orange-400' },
};

export function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [settlements, setSettlements] = useState<Settlement[]>([]);
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState<'holdings' | 'payments' | 'settlements'>('holdings');
  const [expandedTypes, setExpandedTypes] = useState<Set<AccountType>>(new Set(['ASSET']));
  const [showBalances, setShowBalances] = useState(true);
  const [merchantId] = useState('00000000-0000-0000-0000-000000000001');

  const loadData = async () => {
    setLoading(true);
    try {
      const [accountsData, paymentsData, settlementsData] = await Promise.all([
        ledgerApi.listAccounts(),
        paymentApi.list(merchantId),
        settlementApi.list(merchantId),
      ]);
      setAccounts(accountsData || []);
      setPayments(paymentsData || []);
      setSettlements(settlementsData || []);
    } catch (err) {
      console.error('Failed to load data:', err);
      setAccounts([]);
      setPayments([]);
      setSettlements([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const accountGroups: AccountGroup[] = (['ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE'] as AccountType[]).map(type => ({
    type,
    accounts: accounts.filter(a => a.type === type),
    totalBalance: accounts.filter(a => a.type === type).reduce((sum, a) => sum + (a.balance || 0), 0),
  }));

  const holdingsSummary: HoldingsSummary = {
    totalAssets: accountGroups.find(g => g.type === 'ASSET')?.totalBalance || 0,
    totalLiabilities: accountGroups.find(g => g.type === 'LIABILITY')?.totalBalance || 0,
    totalEquity: accountGroups.find(g => g.type === 'EQUITY')?.totalBalance || 0,
    netPosition: (accountGroups.find(g => g.type === 'ASSET')?.totalBalance || 0) - 
                 (accountGroups.find(g => g.type === 'LIABILITY')?.totalBalance || 0),
  };

  const formatAmount = (cents: number, currency: string = 'USD') => {
    const value = showBalances ? cents / 100 : 0;
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
    }).format(value);
  };

  const toggleType = (type: AccountType) => {
    const newExpanded = new Set(expandedTypes);
    if (newExpanded.has(type)) {
      newExpanded.delete(type);
    } else {
      newExpanded.add(type);
    }
    setExpandedTypes(newExpanded);
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
      PENDING: 'secondary',
      AUTHORIZED: 'default',
      CAPTURED: 'default',
      COMPLETED: 'default',
      PROCESSING: 'secondary',
      FAILED: 'destructive',
      CANCELLED: 'outline',
      REFUNDED: 'outline',
    };
    return <Badge variant={variants[status] || 'outline'}>{status}</Badge>;
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Accounts & Holdings</h1>
          <p className="text-slate-500">Manage accounts and view holdings by asset type</p>
        </div>
        <div className="flex gap-2 items-center">
          <Button 
            variant="outline" 
            size="sm"
            onClick={() => setShowBalances(!showBalances)}
            className="border-slate-600 text-slate-600 hover:bg-slate-100"
          >
            {showBalances ? <EyeOff className="h-4 w-4 mr-2" /> : <Eye className="h-4 w-4 mr-2" />}
            {showBalances ? 'Hide' : 'Show'} Balances
          </Button>
          <Button variant="outline" onClick={loadData} disabled={loading} className="border-slate-600 text-slate-600 hover:bg-slate-100">
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Holdings Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="pb-2">
            <CardDescription className="text-slate-400">Total Assets</CardDescription>
            <CardTitle className="text-2xl text-yellow-400">
              {formatAmount(holdingsSummary.totalAssets)}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 text-slate-400 text-sm">
              <Banknote className="h-4 w-4" />
              <span>Current holdings</span>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="pb-2">
            <CardDescription className="text-slate-400">Total Liabilities</CardDescription>
            <CardTitle className="text-2xl text-red-400">
              {formatAmount(holdingsSummary.totalLiabilities)}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 text-slate-400 text-sm">
              <Building2 className="h-4 w-4" />
              <span>Outstanding obligations</span>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="pb-2">
            <CardDescription className="text-slate-400">Net Position</CardDescription>
            <CardTitle className={`text-2xl ${holdingsSummary.netPosition >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              {formatAmount(holdingsSummary.netPosition)}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 text-slate-400 text-sm">
              <PieChart className="h-4 w-4" />
              <span>Assets - Liabilities</span>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="pb-2">
            <CardDescription className="text-slate-400">Total Revenue</CardDescription>
            <CardTitle className="text-2xl text-green-400">
              {formatAmount(accountGroups.find(g => g.type === 'REVENUE')?.totalBalance || 0)}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 text-slate-400 text-sm">
              <TrendingUp className="h-4 w-4" />
              <span>Income to date</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* View Toggle */}
      <div className="flex gap-2">
        <Button 
          variant={view === 'holdings' ? 'default' : 'outline'} 
          onClick={() => setView('holdings')}
          className={view === 'holdings' ? 'bg-yellow-400 text-slate-900 hover:bg-yellow-500' : 'border-slate-600 text-slate-600 hover:bg-slate-100'}
        >
          <Wallet className="h-4 w-4 mr-2" />
          Account Holdings
        </Button>
        <Button 
          variant={view === 'payments' ? 'default' : 'outline'} 
          onClick={() => setView('payments')}
          className={view === 'payments' ? 'bg-yellow-400 text-slate-900 hover:bg-yellow-500' : 'border-slate-600 text-slate-600 hover:bg-slate-100'}
        >
          <ArrowDownLeft className="h-4 w-4 mr-2" />
          Payments (Incoming)
        </Button>
        <Button 
          variant={view === 'settlements' ? 'default' : 'outline'} 
          onClick={() => setView('settlements')}
          className={view === 'settlements' ? 'bg-yellow-400 text-slate-900 hover:bg-yellow-500' : 'border-slate-600 text-slate-600 hover:bg-slate-100'}
        >
          <ArrowUpRight className="h-4 w-4 mr-2" />
          Settlements (Outgoing)
        </Button>
      </div>

      {/* Holdings by Account Type */}
      {view === 'holdings' && (
        <div className="space-y-4">
          {accountGroups.map(group => {
            const config = ACCOUNT_TYPE_CONFIG[group.type];
            const Icon = config.icon;
            const isExpanded = expandedTypes.has(group.type);
            
            return (
              <Card key={group.type} className="border-slate-200">
                <CardHeader 
                  className="cursor-pointer hover:bg-slate-50 transition-colors"
                  onClick={() => toggleType(group.type)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-lg bg-slate-100 ${config.color}`}>
                        <Icon className="h-5 w-5" />
                      </div>
                      <div>
                        <CardTitle className="text-lg">{config.label}</CardTitle>
                        <CardDescription>
                          {group.accounts.length} account{group.accounts.length !== 1 ? 's' : ''}
                        </CardDescription>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="text-right">
                        <p className="text-sm text-slate-500">Total</p>
                        <p className={`text-xl font-bold ${config.color}`}>
                          {formatAmount(group.totalBalance)}
                        </p>
                      </div>
                      {isExpanded ? (
                        <ChevronDown className="h-5 w-5 text-slate-400" />
                      ) : (
                        <ChevronRight className="h-5 w-5 text-slate-400" />
                      )}
                    </div>
                  </div>
                </CardHeader>
                
                {isExpanded && (
                  <CardContent className="pt-0">
                    <Table>
                      <TableHeader>
                        <TableRow className="hover:bg-transparent">
                          <TableHead className="w-24">Code</TableHead>
                          <TableHead>Account Name</TableHead>
                          <TableHead>Type</TableHead>
                          <TableHead className="text-right">Balance</TableHead>
                          <TableHead>Currency</TableHead>
                          <TableHead>Created</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {group.accounts.map(account => (
                          <TableRow key={account.id}>
                            <TableCell className="font-mono text-slate-600">
                              {account.code || account.id.slice(0, 8)}
                            </TableCell>
                            <TableCell className="font-medium">{account.name}</TableCell>
                            <TableCell>
                              <Badge variant="outline" className="border-slate-300">
                                {account.type}
                              </Badge>
                            </TableCell>
                            <TableCell className={`text-right font-medium ${account.balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                              {showBalances ? formatAmount(account.balance, account.currency) : '****'}
                            </TableCell>
                            <TableCell>{account.currency}</TableCell>
                            <TableCell className="text-slate-500">
                              {new Date(account.created_at).toLocaleDateString()}
                            </TableCell>
                          </TableRow>
                        ))}
                        {group.accounts.length === 0 && (
                          <TableRow>
                            <TableCell colSpan={6} className="text-center py-4 text-slate-500">
                              No accounts in this category
                            </TableCell>
                          </TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </CardContent>
                )}
              </Card>
            );
          })}
        </div>
      )}

      {/* Payments View */}
      {view === 'payments' && (
        <Card>
          <CardHeader>
            <CardTitle>Payment Holdings</CardTitle>
            <CardDescription>Incoming payments and transactions</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Capture Method</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Updated</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {payments.map(payment => (
                  <TableRow key={payment.id}>
                    <TableCell className="font-mono text-xs text-slate-600">
                      {payment.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell className="font-medium text-green-600">
                      {formatAmount(payment.amount, payment.currency)}
                    </TableCell>
                    <TableCell>{getStatusBadge(payment.status)}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{payment.capture_method}</Badge>
                    </TableCell>
                    <TableCell className="text-slate-500">
                      {new Date(payment.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-slate-500">
                      {new Date(payment.updated_at).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))}
                {payments.length === 0 && !loading && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-slate-500">
                      No payments found
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Settlements View */}
      {view === 'settlements' && (
        <Card>
          <CardHeader>
            <CardTitle>Settlement Holdings</CardTitle>
            <CardDescription>Outgoing settlements and payouts</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Bank Account</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Completed</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {settlements.map(settlement => (
                  <TableRow key={settlement.id}>
                    <TableCell className="font-mono text-xs text-slate-600">
                      {settlement.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{settlement.type}</Badge>
                    </TableCell>
                    <TableCell className="font-medium text-red-600">
                      {formatAmount(settlement.amount, settlement.currency)}
                    </TableCell>
                    <TableCell>{getStatusBadge(settlement.status)}</TableCell>
                    <TableCell className="font-mono text-xs">
                      ****{settlement.bank_account.slice(-4)}
                    </TableCell>
                    <TableCell className="text-slate-500">
                      {new Date(settlement.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-slate-500">
                      {settlement.completed_at ? new Date(settlement.completed_at).toLocaleString() : '-'}
                    </TableCell>
                  </TableRow>
                ))}
                {settlements.length === 0 && !loading && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8 text-slate-500">
                      No settlements found
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
