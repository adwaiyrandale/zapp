import { useState, useEffect } from 'react';
import { ledgerApi } from '../api/client';
import type { Account, Journal } from '../types';
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
import { RefreshCw, Wallet, FileText } from 'lucide-react';

export function LedgerPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [journals, setJournals] = useState<Journal[]>([]);
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState<'accounts' | 'journals'>('accounts');

  const loadData = async () => {
    setLoading(true);
    try {
      const [accountsData, journalsData] = await Promise.all([
        ledgerApi.listAccounts(),
        ledgerApi.listJournals(),
      ]);
      setAccounts(accountsData || []);
      setJournals(journalsData || []);
    } catch (err) {
      console.error('Failed to load ledger data:', err);
      setAccounts([]);
      setJournals([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const formatAmount = (cents: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
    }).format(cents / 100);
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Ledger</h1>
          <p className="text-slate-500">Double-entry bookkeeping</p>
        </div>
        <div className="flex gap-2">
          <Button 
            variant={view === 'accounts' ? 'default' : 'outline'} 
            onClick={() => setView('accounts')}
            className={view === 'accounts' ? 'bg-yellow-400 text-slate-900 hover:bg-yellow-500' : 'border-slate-600 text-slate-600 hover:bg-slate-100'}
          >
            <Wallet className="h-4 w-4 mr-2" />
            Accounts
          </Button>
          <Button 
            variant={view === 'journals' ? 'default' : 'outline'} 
            onClick={() => setView('journals')}
            className={view === 'journals' ? 'bg-yellow-400 text-slate-900 hover:bg-yellow-500' : 'border-slate-600 text-slate-600 hover:bg-slate-100'}
          >
            <FileText className="h-4 w-4 mr-2" />
            Journals
          </Button>
          <Button variant="outline" onClick={loadData} disabled={loading} className="border-slate-600 text-slate-600 hover:bg-slate-100">
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {view === 'accounts' && (
        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle>Accounts</CardTitle>
            <CardDescription>Chart of accounts with balances</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Balance</TableHead>
                  <TableHead>Currency</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell className="font-medium">{account.name}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className="border-slate-300">{account.type}</Badge>
                    </TableCell>
                    <TableCell className="font-medium">
                      {formatAmount(account.balance, account.currency)}
                    </TableCell>
                    <TableCell>{account.currency}</TableCell>
                    <TableCell className="text-slate-500">
                      {new Date(account.created_at).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
                {accounts.length === 0 && !loading && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-8 text-slate-500">
                      No accounts found
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {view === 'journals' && (
        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle>Journal Entries</CardTitle>
            <CardDescription>All double-entry transactions</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Entries</TableHead>
                  <TableHead>Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {journals.map((journal) => (
                  <TableRow key={journal.id}>
                    <TableCell className="font-mono text-xs text-slate-600">
                      {journal.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>{journal.description}</TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        {journal.lines?.map((line, idx) => (
                          <Badge key={idx} variant="secondary">
                            Dr: {formatAmount(line.debit)} / Cr: {formatAmount(line.credit)}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-slate-500">
                      {new Date(journal.created_at).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
                {journals.length === 0 && !loading && (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center py-8 text-slate-500">
                      No journal entries found
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
