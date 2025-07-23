import type { Account, CreateTransactionRequest, Transaction } from '$lib/types/account';
import { api } from '$lib/services/api';


class AccountStore {
    accounts = $state<Account[] | null>(null);
    isLoading = $state(false);
    error = $state < string | null > (null);

    constructor() { }

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
            const accounts = await api.getAccounts();
            this.accounts = accounts;
            return this.accounts;
        } catch (err) {
            this.error = err instanceof Error ? err.message : 'Failed to fetch accounts';
            throw err;
        } finally {
            this.isLoading = false;
        }
    }

    async fetchAccountByAccountNumber(accountNumber: number): Promise<Account | null> {
        this.isLoading = true;
        this.error = null;

        try {
            const account = await api.getAccount(accountNumber);
            return account;
        } catch (err) {
            this.error = err instanceof Error ? err.message : 'Failed to fetch account';
            throw err;
        } finally {
            this.isLoading = false;
        }
    }

    async deleteAccount(accountNumber: number): Promise<{ success: boolean }> {
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

    async createTransaction(accountNumber: number, amount: number, type: 'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER_CREDIT' | 'TRANSFER_DEBIT'): Promise<void> {
        this.isLoading = true;
        this.error = null;
        try {
            await api.createTransaction({
                accountId: accountNumber.toString(),
                amount: amount,
                transactionType: type,
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
