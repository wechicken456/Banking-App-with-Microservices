import { api, ApiError } from './api';
import type { LoginCredentials, RegisterCredentials, User } from '$lib/types/auth';

export class AuthService {
    private static readonly ACCESS_TOKEN_KEY = 'access_token';
    private static readonly REFRESH_TOKEN_KEY = 'refresh_token';

    static getAccessToken(): string | null {
        if (typeof window === 'undefined') return null;
        return sessionStorage.getItem(this.ACCESS_TOKEN_KEY);
    }

    static getRefreshToken(): string | null {
        if (typeof window === 'undefined') return null;
        return sessionStorage.getItem(this.REFRESH_TOKEN_KEY);
    }

    static setAccessToken(token: string): void {
        if (typeof window === 'undefined') return;
        sessionStorage.setItem(this.ACCESS_TOKEN_KEY, token);
    }

    static setRefreshToken(token: string): void {
        if (typeof window === 'undefined') return;
        sessionStorage.setItem(this.REFRESH_TOKEN_KEY, token);
    }

    static removeAccessToken(): void {
        if (typeof window === 'undefined') return;
        sessionStorage.removeItem(this.ACCESS_TOKEN_KEY);
    }

    static async login(credentials: LoginCredentials): Promise<User> {
        try {
            const response = await api.login(credentials);
            this.setAccessToken(response.accessToken);
            return response.user;
        } catch (error) {
            if (error instanceof ApiError) {
                throw error;
            }
            throw new ApiError('Login failed', 500);
        }
    }

    static async register(userData: RegisterCredentials): Promise<User> {
        try {
            const response = await api.register(userData);
            return response.user;
        } catch (error) {
            if (error instanceof ApiError) {
                throw error;
            }
            throw new ApiError('Registration failed', 500);
        }
    }

    static async renewToken(): Promise<string> {
        try {
            const response = await api.renewToken();
            this.setAccessToken(response.accessToken);
            return response.accessToken;
        } catch (error) {
            this.removeAccessToken();
            throw error;
        }
    }

    static async logout(): Promise<void> {
        await api.logout();
        this.removeAccessToken();
    }

    static isAuthenticated(): boolean {
        return !!this.getAccessToken();
    }
}