export interface Account {
    accountId: string;
    accountNumber: number;
    balance: number;
}

export interface Transaction {
    transactionId: string;
    accountId: string;
    amount: number;
    timestamp: number;
    transactionType: string;
    status: string;
    transferId?: string;
}

export interface CreateTransactionRequest {
    accountId: string;
    amount: number;
    transactionType: 'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER_CREDIT' | 'TRANSFER_DEBIT';
}


