import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import AdminRoute from './components/AdminRoute';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import Toast from './components/Toast';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider } from './context/ToastContext';
import { WebSocketProvider } from './context/WebSocketContext';
import ChatPage from './pages/ChatPage';
import LoginPage from './pages/LoginPage';
import MemberProfilePage from './pages/MemberProfilePage';
import MembersPage from './pages/MembersPage';
import ProfilePage from './pages/ProfilePage';
import SessionCreatePage from './pages/SessionCreatePage';
import SessionDetailPage from './pages/SessionDetailPage';
import SessionEditPage from './pages/SessionEditPage';
import SessionsPage from './pages/SessionsPage';
import VerifyPage from './pages/VerifyPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: true,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <WebSocketProvider>
          <ToastProvider>
            <Toast />
            <BrowserRouter>
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/auth/verify" element={<VerifyPage />} />
                <Route element={<ProtectedRoute />}>
                  <Route element={<Layout />}>
                    <Route path="/" element={<SessionsPage />} />
                    <Route path="/sessions/:id" element={<SessionDetailPage />} />
                    <Route path="/chat" element={<ChatPage />} />
                    <Route path="/chat/:channelId" element={<ChatPage />} />
                    <Route path="/profile" element={<ProfilePage />} />
                    <Route path="/profile/:id" element={<MemberProfilePage />} />
                    <Route element={<AdminRoute />}>
                      <Route path="/sessions/new" element={<SessionCreatePage />} />
                      <Route path="/sessions/:id/edit" element={<SessionEditPage />} />
                      <Route path="/members" element={<MembersPage />} />
                    </Route>
                  </Route>
                </Route>
              </Routes>
            </BrowserRouter>
          </ToastProvider>
        </WebSocketProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}
