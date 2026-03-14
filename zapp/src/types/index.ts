export interface Payment {
  id: string;
  merchant_id: string;
  amount: number;
  currency: string;
  status: PaymentStatus;
  capture_method: CaptureMethod;
  metadata?: string;
  created_at: string;
  updated_at: string;
}

export type PaymentStatus = 'PENDING' | 'AUTHORIZED' | 'CAPTURED' | 'CANCELLED' | 'REFUNDED';
export type CaptureMethod = 'AUTOMATIC' | 'MANUAL';

export interface Charge {
  id: string;
  payment_id: string;
  kind: string;
  amount: number;
  currency: string;
  status: string;
  processor_ref?: string;
  failure_code?: string;
  failure_message?: string;
  created_at: string;
  completed_at?: string;
}

export interface Settlement {
  id: string;
  merchant_id: string;
  payment_id?: string;
  amount: number;
  currency: string;
  type: 'ACH' | 'WIRE';
  status: SettlementStatus;
  bank_account: string;
  routing_number: string;
  trace_number?: string;
  failure_code?: string;
  failure_message?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export type SettlementStatus = 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';

export interface Account {
  id: string;
  name: string;
  type: string;
  balance: number;
  currency: string;
  created_at: string;
  updated_at: string;
}

export interface Journal {
  id: string;
  description: string;
  created_at: string;
  lines: JournalLine[];
}

export interface JournalLine {
  id: string;
  account_id: string;
  debit: number;
  credit: number;
}

export interface CreatePaymentRequest {
  merchant_id: string;
  amount: number;
  currency: string;
  capture_method: string;
  metadata?: string;
}

export interface CreateSettlementRequest {
  merchant_id: string;
  payment_id?: string;
  amount: number;
  currency: string;
  type: 'ACH' | 'WIRE';
  bank_account: string;
  routing_number: string;
}
