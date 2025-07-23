// JavaScript's Number type uses 64-bit floating point, which can only safely represent integers up to Number.MAX_SAFE_INTEGER (2^53 - 1 = 9,007,199,254,740,991). 
// Our Go int64 values are much larger, causing precision loss. 
// Therefore, we use strings to represent large accoutn numbers in TypeScript.
export interface Account {
    accountId: string;
    accountNumber: string;
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


