import { useState, useEffect } from 'react';
import { settlementApi } from '../api/client';
import type { Settlement } from '../types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Badge } from '../components/ui/badge';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '../components/ui/table';
import { Plus, RefreshCw, Play, CheckCircle, XCircle } from 'lucide-react';

const DEMO_MERCHANT_ID = '00000000-0000-0000-0000-000000000001';

const statusVariants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  PENDING: 'secondary',
  PROCESSING: 'secondary',
  COMPLETED: 'default',
  FAILED: 'destructive',
  CANCELLED: 'destructive',
};

export function SettlementsPage() {
  const [settlements, setSettlements] = useState<Settlement[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('USD');
  const [bankAccount, setBankAccount] = useState('');
  const [routingNumber, setRoutingNumber] = useState('');

  const loadSettlements = async () => {
    setLoading(true);
    try {
      const data = await settlementApi.list(DEMO_MERCHANT_ID);
      setSettlements(data || []);
    } catch (err) {
      console.error('Failed to load settlements:', err);
      setSettlements([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSettlements();
  }, []);

  const handleCreate = async () => {
    try {
      await settlementApi.create({
        merchant_id: DEMO_MERCHANT_ID,
        amount: parseInt(amount) * 100,
        currency,
        type: 'ACH',
        bank_account: bankAccount,
        routing_number: routingNumber,
      });
      setAmount('');
      setBankAccount('');
      setRoutingNumber('');
      setShowCreate(false);
      loadSettlements();
    } catch (err) {
      console.error('Failed to create settlement:', err);
    }
  };

  const handleProcess = async (id: string) => {
    try {
      await settlementApi.process(id);
      loadSettlements();
    } catch (err) {
      console.error('Failed to process:', err);
    }
  };

  const handleComplete = async (id: string) => {
    try {
      await settlementApi.complete(id);
      loadSettlements();
    } catch (err) {
      console.error('Failed to complete:', err);
    }
  };

  const handleCancel = async (id: string) => {
    try {
      await settlementApi.cancel(id);
      loadSettlements();
    } catch (err) {
      console.error('Failed to cancel:', err);
    }
  };

  const formatAmount = (cents: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
    }).format(cents / 100);
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Settlements</h1>
          <p className="text-slate-500">Manage ACH and wire transfers</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={loadSettlements} disabled={loading} className="border-slate-600 text-slate-600 hover:bg-slate-100">
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={() => setShowCreate(!showCreate)} className="bg-yellow-400 text-slate-900 hover:bg-yellow-500">
            <Plus className="h-4 w-4 mr-2" />
            New Settlement
          </Button>
        </div>
      </div>

      {showCreate && (
        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle>Create Settlement</CardTitle>
            <CardDescription>Create a new ACH settlement</CardDescription>
          </CardHeader>
          <CardContent className="flex gap-4 flex-wrap">
            <Input
              placeholder="Amount (cents)"
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="w-32"
            />
            <Input
              placeholder="Currency"
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              className="w-24"
            />
            <Input
              placeholder="Bank Account"
              value={bankAccount}
              onChange={(e) => setBankAccount(e.target.value)}
              className="w-40"
            />
            <Input
              placeholder="Routing Number"
              value={routingNumber}
              onChange={(e) => setRoutingNumber(e.target.value)}
              className="w-40"
            />
            <Button onClick={handleCreate} className="bg-yellow-400 text-slate-900 hover:bg-yellow-500">Create</Button>
          </CardContent>
        </Card>
      )}

      <Card className="border-slate-200">
        <CardHeader>
          <CardTitle>Recent Settlements</CardTitle>
          <CardDescription>All settlement transactions</CardDescription>
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
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {settlements.map((settlement) => (
                <TableRow key={settlement.id}>
                  <TableCell className="font-mono text-xs text-slate-600">
                    {settlement.id.slice(0, 8)}...
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{settlement.type}</Badge>
                  </TableCell>
                  <TableCell className="font-medium text-red-600">
                    {formatAmount(settlement.amount)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariants[settlement.status]}>
                      {settlement.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    ****{settlement.bank_account.slice(-4)}
                  </TableCell>
                  <TableCell className="text-slate-500">
                    {new Date(settlement.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      {settlement.status === 'PENDING' && (
                        <Button 
                          size="sm" 
                          variant="outline"
                          onClick={() => handleProcess(settlement.id)}
                          className="border-slate-600 text-slate-600 hover:bg-slate-100"
                        >
                          <Play className="h-3 w-3 mr-1" />
                          Process
                        </Button>
                      )}
                      {settlement.status === 'PROCESSING' && (
                        <Button 
                          size="sm" 
                          variant="outline"
                          onClick={() => handleComplete(settlement.id)}
                          className="border-slate-600 text-slate-600 hover:bg-slate-100"
                        >
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Complete
                        </Button>
                      )}
                      {settlement.status === 'PENDING' && (
                        <Button 
                          size="sm" 
                          variant="destructive"
                          onClick={() => handleCancel(settlement.id)}
                        >
                          <XCircle className="h-3 w-3 mr-1" />
                          Cancel
                        </Button>
                      )}
                    </div>
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
    </div>
  );
}
