import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import type { User, AuthTokens } from '$lib/types/auth';
import { api } from '$lib/services/api';

class AuthStore {
    private readonly ACCESS_TOKEN_KEY = 'accessToken';
    private readonly REFRESH_TOKEN_KEY = 'refreshToken';
    user = $state<User | null>(null);
    isLoading = $state(false);
    isAuthenticated = $derived(this.user !== null);

    constructor() {
        if (browser) {
            this.checkAuth();
            // set the token provider for protected API calls
            api.setTokenProvider({
                getAccessToken: () => this.getAccessToken(),
            });
        }
    }

    private getAccessToken(): string | null {
        if (typeof window === 'undefined') return null;
        return sessionStorage.getItem(this.ACCESS_TOKEN_KEY);
    }

    private getRefreshToken(): string | null {
        if (typeof window === 'undefined') return null;
        return sessionStorage.getItem(this.REFRESH_TOKEN_KEY);
    }

    private setAccessToken(token: string): void {
        if (typeof window === 'undefined') return;
        sessionStorage.setItem(this.ACCESS_TOKEN_KEY, token);
    }

    private setRefreshToken(token: string): void {
        if (typeof window === 'undefined') return;
        sessionStorage.setItem(this.REFRESH_TOKEN_KEY, token);
    }

    private removeAccessToken(): void {
        if (typeof window === 'undefined') return;
        sessionStorage.removeItem(this.ACCESS_TOKEN_KEY);
    }


    async checkAuth() {
        const accessToken = this.getAccessToken();
        if (!accessToken) {
            this.user = null;
            return;
        }

        try {
            this.isLoading = true;
            const user = await api.getProfile();
            this.user = user;
        } catch (error) {
            console.error('Auth check failed:', error);
            this.user = null;
            this.clearTokens();
        } finally {
            this.isLoading = false;
        }
    }

    private getTokens(): AuthTokens | null {
        const accessToken = this.getAccessToken();
        if (!accessToken) {
            console.warn('No access token found, cannot renew');
            return null;
        }
        const refreshToken = this.getRefreshToken();
        if (!refreshToken) {
            console.warn('No refresh token found, cannot renew');
            return null;
        }
        return {
            accessToken,
            refreshToken,
        };
    }

    async login(email: string, password: string) {
        try {
            this.isLoading = true;
            const response = await api.login({ email, password });
            console.log('Login response:', response);

            this.setAccessToken(response.accessToken);
            this.setRefreshToken(response.refreshToken);

            this.user = { id: response.userId, email: response.email };

            goto('/dashboard');
            return { success: true };
        } catch (error) {
            console.error('Login failed:', error);
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Login failed'
            };
        } finally {
            this.isLoading = false;
        }
    }

    async register(email: string, password: string, confirmPassword: string) {
        if (password !== confirmPassword) {
            return { success: false, error: 'Passwords do not match' };
        }

        try {
            this.isLoading = true;
            await api.register({ email, password });
            return { success: true };
        } catch (error) {
            console.error('Registration failed:', error);
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Registration failed'
            };
        } finally {
            this.isLoading = false;
        }
    }

    async logout() {
        try {
            await api.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            this.user = null;
            this.clearTokens();
            goto('/login');
        }
    }

    async renewToken() {
        try {
            const tokens = this.getTokens();
            if (tokens === null) {
                console.warn('One of access token, refresh token, or fingerprint is missing, cannot renew');
                return false;
            }
            await api.renewToken(tokens);
            // Token is renewed via cookies, no need to handle response
            return true;
        } catch (error) {
            console.error('Token renewal failed:', error);
            this.logout();
            return false;
        }
    }

    public clearTokens() {
        if (browser) {
            sessionStorage.removeItem(this.ACCESS_TOKEN_KEY);
            sessionStorage.removeItem(this.REFRESH_TOKEN_KEY);
            // TODO: clear cookies across tabs
        }
    }
}

export const authStore = new AuthStore();
