import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { AuthGuard } from '@/features/auth/components/AuthGuard';

// Lazy loading des pages
const HomePage = lazy(() => import('@/app/HomePage'));
const DashboardPage = lazy(() => import('@/app/pages/DashboardPage'));
const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage'));
const RegisterPage = lazy(() => import('@/features/auth/pages/RegisterPage'));
const TracksPage = lazy(() => import('@/features/tracks/pages/TracksPage'));
const ChatPage = lazy(() => import('@/features/chat/pages/ChatPage'));
const NotFoundPage = lazy(() => import('@/app/NotFoundPage'));

// Composant de fallback pour le loading
const LoadingFallback = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="flex flex-col items-center space-y-4">
      <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      <p className="text-gray-600">Chargement...</p>
    </div>
  </div>
);

export const Router = () => (
  <Suspense fallback={<LoadingFallback />}>
    <Routes>
      {/* Redirect root to dashboard */}
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      
      {/* Pages publiques (avec redirection si déjà connecté) */}
      <Route path="/login" element={
        <AuthGuard requireAuth={false}>
          <LoginPage />
        </AuthGuard>
      } />
      <Route path="/register" element={
        <AuthGuard requireAuth={false}>
          <RegisterPage />
        </AuthGuard>
      } />
      
      {/* Pages protégées (authentification requise) */}
      <Route path="/dashboard" element={
        <AuthGuard requireAuth={true}>
          <DashboardPage />
        </AuthGuard>
      } />
      <Route path="/tracks" element={
        <AuthGuard requireAuth={true}>
          <TracksPage />
        </AuthGuard>
      } />
      <Route path="/chat" element={
        <AuthGuard requireAuth={true}>
          <ChatPage />
        </AuthGuard>
      } />
      <Route path="/home" element={
        <AuthGuard requireAuth={true}>
          <HomePage />
        </AuthGuard>
      } />
      
      {/* 404 */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  </Suspense>
); 