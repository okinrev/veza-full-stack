import { Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './Layout';
import { AuthGuard } from '@/features/auth/components/AuthGuard';
import { LoginPage } from '@/features/auth/pages/LoginPage';
import { RegisterPage } from '@/features/auth/pages/RegisterPage';
import { ChatPage } from '@/features/chat/pages/ChatPage';

// Page simple de dashboard temporaire avec navigation vers chat
const DashboardPage = () => (
  <div className="min-h-screen bg-gray-50 p-6">
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">🎶 Tableau de bord Veza</h1>
      <div className="bg-white rounded-lg shadow p-6">
        <p className="text-gray-600 mb-4">Bienvenue dans votre espace Veza !</p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg">
            <h3 className="font-semibold text-blue-900">🎵 Mes Pistes</h3>
            <p className="text-blue-700 text-sm">Gérez vos pistes audio</p>
          </div>
          <a href="/chat" className="bg-purple-50 p-4 rounded-lg hover:bg-purple-100 transition-colors">
            <h3 className="font-semibold text-purple-900">💬 Chat</h3>
            <p className="text-purple-700 text-sm">Communiquez avec la communauté</p>
          </a>
          <div className="bg-green-50 p-4 rounded-lg">
            <h3 className="font-semibold text-green-900">📁 Ressources</h3>
            <p className="text-green-700 text-sm">Partagez vos ressources</p>
          </div>
        </div>
      </div>
    </div>
  </div>
);

// Page 404 simple
const NotFoundPage = () => (
  <div className="min-h-screen bg-gray-50 flex items-center justify-center">
    <div className="text-center">
      <h1 className="text-6xl font-bold text-gray-900 mb-4">404</h1>
      <p className="text-xl text-gray-600 mb-8">Page non trouvée</p>
      <a 
        href="/dashboard" 
        className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
      >
        Retour au tableau de bord
      </a>
    </div>
  </div>
);

export const Router = () => (
  <Routes>
    {/* Redirect root to dashboard */}
    <Route path="/" element={<Navigate to="/dashboard" replace />} />
    
    {/* Pages publiques (sans navbar, avec redirection si déjà connecté) */}
    <Route path="/login" element={
      <AuthGuard requireAuth={false}>
        <Layout showNavbar={false}>
          <LoginPage />
        </Layout>
      </AuthGuard>
    } />
    <Route path="/register" element={
      <AuthGuard requireAuth={false}>
        <Layout showNavbar={false}>
          <RegisterPage />
        </Layout>
      </AuthGuard>
    } />
    
    {/* Pages protégées (avec navbar, authentification requise) */}
    <Route path="/dashboard" element={
      <AuthGuard requireAuth={true}>
        <Layout>
          <DashboardPage />
        </Layout>
      </AuthGuard>
    } />
    
    <Route path="/chat" element={
      <AuthGuard requireAuth={true}>
        <Layout>
          <ChatPage />
        </Layout>
      </AuthGuard>
    } />
    
    {/* 404 */}
    <Route path="*" element={
      <Layout showNavbar={false}>
        <NotFoundPage />
      </Layout>
    } />
  </Routes>
); 