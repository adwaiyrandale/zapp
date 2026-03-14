import type { Payment, Settlement, Account, Journal, CreatePaymentRequest, CreateSettlementRequest, Charge } from '../types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${endpoint}`;
  
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.json();
}

export const paymentApi = {
  list: (merchantId: string) => 
    request<Payment[]>(`/api/v1/payments?merchant_id=${merchantId}`),
  
  get: (id: string) => 
    request<Payment>(`/api/v1/payments/${id}`),
  
  create: (data: CreatePaymentRequest) => 
    request<Payment>('/api/v1/payments', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  
  authorize: (id: string) => 
    request<{ payment: Payment; charge: Charge }>(`/api/v1/payments/${id}/authorize`, {
      method: 'POST',
    }),
  
  capture: (id: string) => 
    request<{ payment: Payment; charge: Charge }>(`/api/v1/payments/${id}/capture`, {
      method: 'POST',
    }),
  
  cancel: (id: string) => 
    request<Payment>(`/api/v1/payments/${id}/cancel`, {
      method: 'POST',
    }),
  
  refund: (id: string, amount?: number) => 
    request<{ payment: Payment; charge: Charge }>(`/api/v1/payments/${id}/refund${amount ? `?amount=${amount}` : ''}`, {
      method: 'POST',
    }),
  
  getCharges: (id: string) => 
    request<Charge[]>(`/api/v1/payments/${id}/charges`),
};

export const settlementApi = {
  list: (merchantId: string) => 
    request<Settlement[]>(`/api/v1/settlements?merchant_id=${merchantId}`),
  
  get: (id: string) => 
    request<Settlement>(`/api/v1/settlements/${id}`),
  
  create: (data: CreateSettlementRequest) => 
    request<Settlement>('/api/v1/settlements', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  
  process: (id: string) => 
    request<Settlement>(`/api/v1/settlements/${id}/process`, {
      method: 'POST',
    }),
  
  complete: (id: string) => 
    request<Settlement>(`/api/v1/settlements/${id}/complete`, {
      method: 'POST',
    }),
  
  cancel: (id: string) => 
    request<Settlement>(`/api/v1/settlements/${id}/cancel`, {
      method: 'POST',
    }),
};

export const ledgerApi = {
  listAccounts: () => 
    request<Account[]>('/api/v1/ledger/accounts'),
  
  getAccount: (id: string) => 
    request<Account>(`/api/v1/ledger/accounts/${id}`),
  
  getBalance: (id: string) => 
    request<{ account_id: string; balance: number; currency: string }>(`/api/v1/ledger/accounts/${id}/balance`),
  
  listJournals: () => 
    request<Journal[]>('/api/v1/ledger/journals'),
  
  getJournal: (id: string) => 
    request<Journal>(`/api/v1/ledger/journals/${id}`),
};

export { ApiError };
