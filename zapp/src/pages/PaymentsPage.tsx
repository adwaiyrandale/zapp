import { useState, useEffect } from 'react';
import { paymentApi } from '../api/client';
import type { Payment } from '../types';
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
import { Plus, RefreshCw, CheckCircle, XCircle } from 'lucide-react';

const DEMO_MERCHANT_ID = '00000000-0000-0000-0000-000000000001';

const statusVariants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  PENDING: 'secondary',
  AUTHORIZED: 'default',
  CAPTURED: 'default',
  CANCELLED: 'destructive',
  REFUNDED: 'outline',
};

export function PaymentsPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('USD');

  const loadPayments = async () => {
    setLoading(true);
    try {
      const data = await paymentApi.list(DEMO_MERCHANT_ID);
      setPayments(data || []);
    } catch (err) {
      console.error('Failed to load payments:', err);
      setPayments([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPayments();
  }, []);

  const handleCreate = async () => {
    try {
      await paymentApi.create({
        merchant_id: DEMO_MERCHANT_ID,
        amount: parseInt(amount) * 100,
        currency,
        capture_method: 'MANUAL',
      });
      setAmount('');
      setShowCreate(false);
      loadPayments();
    } catch (err) {
      console.error('Failed to create payment:', err);
    }
  };

  const handleAuthorize = async (id: string) => {
    try {
      await paymentApi.authorize(id);
      loadPayments();
    } catch (err) {
      console.error('Failed to authorize:', err);
    }
  };

  const handleCapture = async (id: string) => {
    try {
      await paymentApi.capture(id);
      loadPayments();
    } catch (err) {
      console.error('Failed to capture:', err);
    }
  };

  const handleCancel = async (id: string) => {
    try {
      await paymentApi.cancel(id);
      loadPayments();
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
          <h1 className="text-3xl font-bold text-slate-900">Payments</h1>
          <p className="text-slate-500">Manage payment transactions</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={loadPayments} disabled={loading} className="border-slate-600 text-slate-600 hover:bg-slate-100">
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={() => setShowCreate(!showCreate)} className="bg-yellow-400 text-slate-900 hover:bg-yellow-500">
            <Plus className="h-4 w-4 mr-2" />
            New Payment
          </Button>
        </div>
      </div>

      {showCreate && (
      <Card className="border-slate-200">
          <CardHeader>
            <CardTitle>Create Payment</CardTitle>
            <CardDescription>Create a new payment intent</CardDescription>
          </CardHeader>
          <CardContent className="flex gap-4">
            <Input
              placeholder="Amount"
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
            <Button onClick={handleCreate} className="bg-yellow-400 text-slate-900 hover:bg-yellow-500">Create</Button>
          </CardContent>
        </Card>
      )}

      <Card className="border-slate-200">
        <CardHeader>
          <CardTitle>Recent Payments</CardTitle>
          <CardDescription>All payment transactions</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Method</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {payments.map((payment) => (
                <TableRow key={payment.id}>
                  <TableCell className="font-mono text-xs text-slate-600">
                    {payment.id.slice(0, 8)}...
                  </TableCell>
                  <TableCell className="font-medium text-green-600">
                    {formatAmount(payment.amount)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariants[payment.status]}>
                      {payment.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{payment.capture_method}</TableCell>
                  <TableCell>
                    {new Date(payment.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      {payment.status === 'PENDING' && (
                        <Button 
                          size="sm" 
                          variant="outline"
                          onClick={() => handleAuthorize(payment.id)}
                          className="border-slate-600 text-slate-600 hover:bg-slate-100"
                        >
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Authorize
                        </Button>
                      )}
                      {payment.status === 'AUTHORIZED' && (
                        <Button 
                          size="sm" 
                          variant="outline"
                          onClick={() => handleCapture(payment.id)}
                          className="border-slate-600 text-slate-600 hover:bg-slate-100"
                        >
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Capture
                        </Button>
                      )}
                      {(payment.status === 'PENDING' || payment.status === 'AUTHORIZED') && (
                        <Button 
                          size="sm" 
                          variant="destructive"
                          onClick={() => handleCancel(payment.id)}
                        >
                          <XCircle className="h-3 w-3 mr-1" />
                          Cancel
                        </Button>
                      )}
                    </div>
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
    </div>
  );
}
