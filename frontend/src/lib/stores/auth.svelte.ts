import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import type { User } from '$lib/types/auth';
import { api } from '$lib/services/api';

class AuthStore {
    user = $state<User | null>(null);
    isLoading = $state(false);
    isAuthenticated = $derived(this.user !== null);

    constructor() {
        if (browser) {
            this.checkAuth();
        }
    }

    private async checkAuth() {
        const accessToken = sessionStorage.get('access_token');
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

    async login(email: string, password: string) {
        try {
            this.isLoading = true;
            const response = await api.login({ email, password });
            
            // Cookies are set by the server via Set-Cookie headers
            // We just need to set the user data
            this.user = response.user;

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
            await api.register({ email, password, confirmPassword });
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

    private clearTokens() {
        if (browser) {
            sessionStorage.remove('access_token');
            sessionStorage.remove('refresh_token');
            sessionStorage.remove('fingerprint');

            // TODO: clear cookies across tabs
        }
    }

    async renewToken() {
        try {
            await api.renewToken();
            // Token is renewed via cookies, no need to handle response
            return true;
        } catch (error) {
            console.error('Token renewal failed:', error);
            this.logout();
            return false;
        }
    }
}

export const authStore = new AuthStore();