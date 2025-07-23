import type { Account, CreateTransactionRequest, Transaction } from '$lib/types/account';
import { api } from '$lib/services/api';


class AccountStore {
    accounts = $state<Account[]>([]);
    isLoading = $state(false);
    error = $state < string | null > (null);
    
    constructor() {}

    async createAccount(balance : number) : Promise<Account> {
        this.isLoading = true;
        this.error = null;

        try {
            const result = await api.createAccount(balance);
            // Refresh accounts list after creating new account
            await this.fetchAllAccounts();
            return result;
        } catch (error) {
            this.error = error instanceof Error ? error.message : 'Failed to create account';
            throw error;
        } finally {
            this.isLoading = false;
        }
    }

    async fetchAllAccounts(): Promise<Account[]> {
        this.isLoading = true;
        this.error = null;

        try {
            const res = await api.getAccounts();
            const accounts = res.accounts;
            if (!accounts || accounts.length == 0) {
                this.accounts = [];
                return [];
            }
            this.accounts = accounts;
            return this.accounts;
        } catch (err) {
            this.accounts = [];
            this.error = err instanceof Error ? err.message : 'Failed to fetch accounts';
            throw err;
        } finally {
            this.isLoading = false;
        }
    }

    async fetchAccountByAccountNumber(accountNumber: string): Promise<Account | null> {
        this.isLoading = true;
        this.error = null;

        try {
            const resp = await api.getAccount(accountNumber);
            return resp.account;
        } catch (err) {
            this.error = err instanceof Error ? err.message : 'Failed to fetch account';
            throw err;
        } finally {
            this.isLoading = false;
        }
    }

    async deleteAccount(accountNumber: string): Promise<{ success: boolean }> {
        this.isLoading = true;
        this.error = null;
        
        try {
            const result = await api.deleteAccount(accountNumber);
            if (result.success) {
                // Refresh accounts list after deletion
                await this.fetchAllAccounts();
            }
            return result;
        } catch (error) {
            this.error = error instanceof Error ? error.message : 'Failed to delete account';
            throw error;
        } finally {
            this.isLoading = false;
        }
    }

    async createTransaction(accountNumber: string, amount: number, type: 'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER_CREDIT' | 'TRANSFER_DEBIT'): Promise<void> {
        this.isLoading = true;
        this.error = null;
        let transactionType: 'CREDIT' | 'DEBIT' | 'TRANSFER_CREDIT' | 'TRANSFER_DEBIT';
        switch  (type) {
            case 'DEPOSIT': {
                transactionType = 'CREDIT';
                break;
            } 
            case 'WITHDRAWAL': {
                transactionType = 'DEBIT';
                break;
            } 
            case 'TRANSFER_CREDIT': {
                transactionType = 'TRANSFER_CREDIT';
                break;
            } 
            case 'TRANSFER_DEBIT': {
                transactionType = 'TRANSFER_DEBIT';
                break;
            }
            default: {
                throw new Error('Invalid transaction type');
            }
        }

        try {
            await api.createTransaction({
                accountId: accountNumber.toString(),
                amount: amount,
                transactionType: transactionType,
            } as CreateTransactionRequest); 
            // Refresh accounts list after transaction
            await this.fetchAllAccounts();
        } catch (error) {
            this.error = error instanceof Error ? error.message : 'Failed to create transaction';
            throw error;
        } finally {
            this.isLoading = false;
        }
    }

    async fetchTransactionsByAccountId(accountId: string): Promise<Transaction[]> {
        this.isLoading = true;
        this.error = null;
        
        try {
            const transactions = await api.getTransactionsByAccountId(accountId);
            return transactions;
        } catch (error) {
            this.error = error instanceof Error ? error.message : 'Failed to fetch transactions';
            throw error;
        } finally {
            this.isLoading = false;
        }
    }

    clearError() {
        this.error = null;
    }
}

export const accountStore = new AccountStore();
